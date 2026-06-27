package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vo0ov/tg2txt/internal/activity"
	"github.com/vo0ov/tg2txt/internal/formatter"
	"github.com/vo0ov/tg2txt/internal/telegram"
)

type Options struct {
	Source              string
	Destination         string
	ActivityDestination string
	SkipService         bool
	SkipHeader          bool
	Format              formatter.Options
	Join                JoinOptions
}

type JoinOptions struct {
	Enabled       bool
	Separator     string
	WindowSeconds int
}

type Result struct {
	Total               int
	Written             int
	Skipped             int
	Destination         string
	ActivityDestination string
}

func Convert(options Options) (result Result, err error) {
	result = Result{
		Destination:         options.Destination,
		ActivityDestination: options.ActivityDestination,
	}

	export, err := loadExport(options.Source)
	if err != nil {
		return result, err
	}

	err = os.MkdirAll(filepath.Dir(options.Destination), 0750)
	if err != nil {
		return result, fmt.Errorf("failed to create output directory: %w", err)
	}

	out, err := os.Create(options.Destination)
	if err != nil {
		return result, fmt.Errorf("failed to create %s: %w", options.Destination, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", options.Destination, closeErr)
		}
	}()

	var lines []outputLine
	if export.Name != "" && !options.SkipHeader {
		lines = append(lines, outputLine{Line: "# " + export.Name}, outputLine{})
	}

	formatOptions := options.Format
	formatOptions.ChatName = export.Name
	formatOptions.ChatType = export.Type
	formatOptions.ChatID = export.ID

	result.Total = len(export.Messages)
	for i := range export.Messages {
		msg := &export.Messages[i]

		if options.SkipService && msg.Type == "service" {
			result.Skipped++
			continue
		}

		formatted, ok := formatter.Format(msg, formatOptions)
		if !ok {
			result.Skipped++
			continue
		}

		line := outputLine{
			Line:      formatted.Line,
			Sender:    formatted.Sender,
			Body:      formatted.Body,
			Mergeable: formatted.Mergeable,
		}
		if when, ok := messageTime(msg); ok {
			line.Time = when
			line.HasTime = true
		}

		if options.Join.Enabled && mergeLastLine(lines, line, options.Join) {
			lines[len(lines)-1].Line += options.Join.Separator + line.Body
			lines[len(lines)-1].Time = line.Time
			lines[len(lines)-1].HasTime = line.HasTime
			continue
		}

		lines = append(lines, line)
	}

	content := strings.Join(outputStrings(lines), "\n") + "\n"
	if _, err = out.WriteString(content); err != nil {
		return result, fmt.Errorf("failed to write output file: %w", err)
	}

	if options.ActivityDestination != "" {
		if _, err := activity.WritePNG(export.Messages, activity.Options{
			Destination: options.ActivityDestination,
			ChatName:    export.Name,
		}); err != nil {
			return result, err
		}
	}

	result.Written = result.Total - result.Skipped
	return result, nil
}

func loadExport(source string) (_ telegram.Export, err error) {
	cleanSource := filepath.Clean(source)
	rootPath := filepath.Dir(cleanSource)
	fileName := filepath.Base(cleanSource)

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return telegram.Export{}, fmt.Errorf("failed to open input directory %s: %w", rootPath, err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close input directory %s: %w", rootPath, closeErr)
		}
	}()

	f, err := root.Open(fileName)
	if err != nil {
		return telegram.Export{}, fmt.Errorf("failed to open %s: %w", source, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", source, closeErr)
		}
	}()

	var export telegram.Export
	err = json.NewDecoder(f).Decode(&export)
	if err != nil {
		return telegram.Export{}, fmt.Errorf("failed to parse Telegram JSON export: %w", err)
	}
	return export, nil
}

type outputLine struct {
	Line      string
	Sender    string
	Body      string
	Time      time.Time
	HasTime   bool
	Mergeable bool
}

func mergeLastLine(lines []outputLine, next outputLine, options JoinOptions) bool {
	if len(lines) == 0 || !next.Mergeable {
		return false
	}

	prev := lines[len(lines)-1]
	if !prev.Mergeable || prev.Sender != next.Sender || !prev.HasTime || !next.HasTime {
		return false
	}

	window := time.Duration(options.WindowSeconds) * time.Second
	if window < 0 {
		return false
	}
	delta := next.Time.Sub(prev.Time)
	return delta >= 0 && delta <= window
}

func messageTime(msg *telegram.Message) (time.Time, bool) {
	t, err := time.Parse("2006-01-02T15:04:05", msg.Date)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func outputStrings(lines []outputLine) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line.Line)
	}
	return out
}
