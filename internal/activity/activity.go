package activity

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/vo0ov/tg2txt/internal/telegram"
	chart "github.com/wcharczuk/go-chart/v2"
)

const telegramTimeLayout = "2006-01-02T15:04:05"

type Options struct {
	Destination string
	ChatName    string
}

type Result struct {
	Destination string
	FirstDay    time.Time
	LastDay     time.Time
	Points      int
}

func WritePNG(messages []telegram.Message, options Options) (result Result, err error) {
	series, err := buildSeries(messages)
	if err != nil {
		return Result{}, err
	}
	xValues, yValues := chartValues(series)

	mkdirErr := os.MkdirAll(filepath.Dir(options.Destination), 0750)
	if mkdirErr != nil {
		return Result{}, fmt.Errorf("failed to create output directory: %w", mkdirErr)
	}

	out, err := os.Create(options.Destination)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create %s: %w", options.Destination, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", options.Destination, closeErr)
		}
	}()

	title := "Telegram activity by day"
	if options.ChatName != "" {
		title = fmt.Sprintf("%s activity by day", options.ChatName)
	}

	graph := chart.Chart{
		Title:  title,
		Width:  1600,
		Height: 900,
		Background: chart.Style{
			FillColor: chart.ColorWhite,
			Padding: chart.Box{
				Top:    48,
				Left:   16,
				Right:  24,
				Bottom: 16,
			},
		},
		Canvas: chart.Style{
			FillColor: chart.ColorLightGray.WithAlpha(24),
		},
		XAxis: chart.XAxis{
			Name:           "Date",
			ValueFormatter: chart.TimeDateValueFormatter,
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray.WithAlpha(48),
				StrokeWidth: 1,
			},
		},
		YAxis: chart.YAxis{
			Name: "Messages",
			ValueFormatter: func(v interface{}) string {
				value, ok := v.(float64)
				if !ok {
					return ""
				}
				return fmt.Sprintf("%d", int(math.Round(value)))
			},
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray.WithAlpha(48),
				StrokeWidth: 1,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Messages",
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 3,
					FillColor:   chart.ColorBlue.WithAlpha(48),
					DotColor:    chart.ColorBlue,
					DotWidth:    4,
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	renderErr := graph.Render(chart.PNG, out)
	if renderErr != nil {
		return Result{}, fmt.Errorf("failed to render %s: %w", options.Destination, renderErr)
	}

	return Result{
		Destination: options.Destination,
		FirstDay:    series.Days[0],
		LastDay:     series.Days[len(series.Days)-1],
		Points:      len(series.Days),
	}, nil
}

type Series struct {
	Days   []time.Time
	Counts []float64
}

func BuildSeries(messages []telegram.Message) (Series, error) {
	return buildSeries(messages)
}

func chartValues(series Series) ([]time.Time, []float64) {
	if len(series.Days) != 1 {
		return series.Days, series.Counts
	}

	return []time.Time{
			series.Days[0],
			series.Days[0].Add(12 * time.Hour),
		}, []float64{
			series.Counts[0],
			series.Counts[0],
		}
}

func buildSeries(messages []telegram.Message) (Series, error) {
	perDay := make(map[time.Time]float64)

	var first time.Time
	var last time.Time
	var found bool

	for _, msg := range messages {
		if msg.Type != "message" {
			continue
		}

		when, err := time.Parse(telegramTimeLayout, msg.Date)
		if err != nil {
			continue
		}

		day := time.Date(when.Year(), when.Month(), when.Day(), 0, 0, 0, 0, time.UTC)
		perDay[day]++

		if !found || day.Before(first) {
			first = day
		}
		if !found || day.After(last) {
			last = day
		}
		found = true
	}

	if !found {
		return Series{}, fmt.Errorf("no dated Telegram messages found for activity chart")
	}

	days := make([]time.Time, 0, int(last.Sub(first).Hours()/24)+1)
	counts := make([]float64, 0, cap(days))
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		days = append(days, day)
		counts = append(counts, perDay[day])
	}

	return Series{
		Days:   days,
		Counts: counts,
	}, nil
}
