package activity

import (
	"testing"
	"time"

	"github.com/vo0ov/tg2txt/internal/telegram"
)

func TestBuildSeriesFillsWholeRangeByDay(t *testing.T) {
	series, err := BuildSeries([]telegram.Message{
		{Type: "message", Date: "2025-09-23T20:51:31"},
		{Type: "message", Date: "2025-09-23T21:51:31"},
		{Type: "service", Date: "2025-09-24T21:51:31"},
		{Type: "message", Date: "2025-09-25T10:00:00"},
	})
	if err != nil {
		t.Fatalf("BuildSeries() error = %v", err)
	}

	if len(series.Days) != 3 {
		t.Fatalf("len(series.Days) = %d, want 3", len(series.Days))
	}

	want := []float64{2, 0, 1}
	for i := range want {
		if series.Counts[i] != want[i] {
			t.Fatalf("series.Counts[%d] = %v, want %v", i, series.Counts[i], want[i])
		}
	}
}

func TestBuildSeriesRequiresAtLeastOneDatedMessage(t *testing.T) {
	_, err := BuildSeries([]telegram.Message{
		{Type: "service", Date: "2025-09-23T20:51:31"},
		{Type: "message", Date: "broken"},
	})
	if err == nil {
		t.Fatal("BuildSeries() error = nil, want non-nil")
	}
}

func TestChartValuesExpandsSingleDaySeries(t *testing.T) {
	daySeries := Series{
		Days:   []time.Time{time.Date(2025, time.September, 23, 0, 0, 0, 0, time.UTC)},
		Counts: []float64{1},
	}

	xValues, yValues := chartValues(daySeries)
	if len(xValues) != 2 {
		t.Fatalf("len(xValues) = %d, want 2", len(xValues))
	}
	if len(yValues) != 2 {
		t.Fatalf("len(yValues) = %d, want 2", len(yValues))
	}
	if !xValues[1].After(xValues[0]) {
		t.Fatal("expected second x value to be after first")
	}
	if yValues[0] != 1 || yValues[1] != 1 {
		t.Fatalf("yValues = %v, want [1 1]", yValues)
	}
}
