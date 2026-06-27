package stats

import (
	"strings"
	"testing"

	"github.com/vo0ov/tg2txt/internal/telegram"
)

func TestRenderIncludesRequestedMetrics(t *testing.T) {
	report := Render(telegram.Export{
		Name: "Team Chat",
		Messages: []telegram.Message{
			{
				ID:   1,
				Type: "message",
				Date: "2025-09-23T09:00:00",
				From: "Alice",
				Text: rawJSON(`"hello world"`),
				Reactions: []telegram.Reaction{
					{Emoji: "👍", Count: 2},
				},
			},
			{
				ID:   2,
				Type: "message",
				Date: "2025-09-23T09:01:00",
				From: "Alice",
				Text: rawJSON(`"again"`),
			},
			{
				ID:               3,
				Type:             "message",
				Date:             "2025-09-23T09:03:00",
				From:             "Bob",
				Text:             rawJSON(`"ok"`),
				MediaType:        "photo",
				ReplyToMessageID: 1,
			},
			{
				ID:     4,
				Type:   "service",
				Date:   "2025-09-24T10:00:00",
				Actor:  "Alice",
				Action: "pin_message",
			},
			{
				ID:            5,
				Type:          "message",
				Date:          "2025-09-25T23:00:00",
				From:          "Alice",
				Text:          rawJSON(`"late"`),
				ForwardedFrom: "Carol",
			},
		},
	})

	assertContains(t, report, "Total messages: 4")
	assertContains(t, report, "Active days: 2")
	assertContains(t, report, "Zero-message days: 1")
	assertContains(t, report, "Max messages/day: 3 on 2025-09-23")
	assertContains(t, report, "Most messages: Alice (3)")
	assertContains(t, report, "Dominance coefficient: 75.00%")
	assertContains(t, report, "Service events: 1")
	assertContains(t, report, "Media messages: 1 (25.00%)")
	assertContains(t, report, "Replies: 1")
	assertContains(t, report, "Forwards: 1")
	assertContains(t, report, "Message Type Top")
	assertContains(t, report, "  text: 3 (60.00%)")
	assertContains(t, report, "  photo: 1 (20.00%)")
	assertContains(t, report, "  service:pin_message: 1 (20.00%)")
	assertContains(t, report, "Days absent while chat was active: 1")
	assertContains(t, report, "Reaction top:\n      👍: 2")
	assertContains(t, report, "  Bob\n    Messages: 1 (25.00%)")
	assertContains(t, report, "Median response: 2m 0s")
	assertContains(t, report, "Media / replies / forwards: 1 / 1 / 0")
}

func rawJSON(s string) []byte {
	return []byte(s)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("report does not contain %q:\n%s", want, got)
	}
}
