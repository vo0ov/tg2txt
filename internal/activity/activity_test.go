package activity

import (
	"testing"

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
