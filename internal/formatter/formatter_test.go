package formatter

import (
	"encoding/json"
	"testing"

	"github.com/vo0ov/tg2txt/internal/telegram"
)

func rawJSON(s string) json.RawMessage {
	return json.RawMessage(s)
}

func TestDefaultOutputIncludesDetailedParts(t *testing.T) {
	msg := &telegram.Message{
		ID:        42,
		Type:      "message",
		Date:      "2025-09-23T20:51:31",
		From:      "Alice",
		Text:      rawJSON(`"hello"`),
		MediaType: "photo",
		Reactions: []telegram.Reaction{
			{
				Emoji: "👍",
				Count: 1,
				Recent: []telegram.ReactionFrom{
					{From: "Bob"},
				},
			},
		},
	}

	got, ok := Message(msg, Options{})
	if !ok {
		t.Fatal("Message returned ok=false")
	}

	want := "[23.09.25 20:51] #42 Alice: hello  [📷 photo]  [Reactions: 👍 by Bob]"
	if got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}

func TestPlainDialogueSkipsEmptyMessages(t *testing.T) {
	msg := &telegram.Message{
		ID:        42,
		Type:      "message",
		Date:      "2025-09-23T20:51:31",
		From:      "Alice",
		MediaType: "photo",
	}

	_, ok := Message(msg, Options{
		SkipTime:      true,
		SkipID:        true,
		SkipMedia:     true,
		SkipReactions: true,
		PlainDialogue: true,
	})
	if ok {
		t.Fatal("Message returned ok=true for empty plain-dialogue message")
	}
}

func TestAnonOnlyAppliesToPersonalChat(t *testing.T) {
	msg := &telegram.Message{
		Type:   "message",
		From:   "Bob",
		FromID: "user123",
		Text:   rawJSON(`"hello"`),
	}

	got, ok := Message(msg, Options{
		SkipTime: true,
		SkipID:   true,
		ChatName: "Bob",
		ChatType: "private_group",
		ChatID:   123,
		AnonPeer: "Peer",
		AnonSelf: "Self",
	})
	if !ok {
		t.Fatal("Message returned ok=false")
	}

	want := "Bob: hello"
	if got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}

func TestPersonalChatAnonUsesPeerFromID(t *testing.T) {
	msg := &telegram.Message{
		Type:   "message",
		From:   "Unknown",
		FromID: "user123",
		Text:   rawJSON(`"hello"`),
	}

	got, ok := Message(msg, Options{
		SkipTime: true,
		SkipID:   true,
		ChatName: "Bob",
		ChatType: "personal_chat",
		ChatID:   123,
		AnonPeer: "Peer",
		AnonSelf: "Self",
	})
	if !ok {
		t.Fatal("Message returned ok=false")
	}

	want := "Peer: hello"
	if got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}

func TestNoEntitiesExtractsPlainText(t *testing.T) {
	msg := &telegram.Message{
		Type: "message",
		From: "Alice",
		Text: rawJSON(`[
			{"type":"bold","text":"bold"},
			" plain ",
			{"type":"code","text":"code"}
		]`),
	}

	got, ok := Message(msg, Options{
		SkipTime:     true,
		SkipID:       true,
		SkipEntities: true,
	})
	if !ok {
		t.Fatal("Message returned ok=false")
	}

	want := "Alice: bold plain code"
	if got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}

func TestNoFlagsRemoveIndependentParts(t *testing.T) {
	msg := &telegram.Message{
		ID:               42,
		Type:             "message",
		Date:             "2025-09-23T20:51:31",
		From:             "Alice",
		Text:             rawJSON(`"hello"`),
		MediaType:        "photo",
		ForwardedFrom:    "Carol",
		ReplyToMessageID: 41,
		Reactions: []telegram.Reaction{
			{Emoji: "👍", Count: 2},
		},
	}

	got, ok := Message(msg, Options{
		SkipTime:      true,
		SkipID:        true,
		SkipMedia:     true,
		SkipReactions: true,
		SkipForwards:  true,
		SkipReplies:   true,
	})
	if !ok {
		t.Fatal("Message returned ok=false")
	}

	want := "Alice: hello"
	if got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}
