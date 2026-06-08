package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vo0ov/tg2txt/internal/telegram"
)

type Options struct {
	SkipTime      bool
	SkipID        bool
	SkipMedia     bool
	SkipReactions bool
	SkipEntities  bool
	SkipForwards  bool
	SkipReplies   bool
	PlainDialogue bool
	AnonPeer      string
	AnonSelf      string
	ChatName      string
	ChatType      string
	ChatID        int64
}

type FormattedMessage struct {
	Line      string
	Sender    string
	Body      string
	Mergeable bool
}

// formatDate converts Telegram timestamps into a compact chat-log timestamp.
func formatDate(raw string) string {
	t, err := time.Parse("2006-01-02T15:04:05", raw)
	if err != nil {
		return raw
	}
	return t.Format("02.01.06 15:04")
}

func parseDuration(raw json.RawMessage) (int, bool) {
	if len(raw) == 0 {
		return 0, false
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, n > 0
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if n, err := strconv.Atoi(s); err == nil {
			return n, n > 0
		}
	}

	return 0, false
}

func formatDuration(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func extractText(raw json.RawMessage, plain bool) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return ""
	}

	var b strings.Builder
	for _, elem := range arr {
		var es string
		if err := json.Unmarshal(elem, &es); err == nil {
			b.WriteString(es)
			continue
		}

		var chunk telegram.TextChunk
		if err := json.Unmarshal(elem, &chunk); err != nil {
			continue
		}

		t := chunk.Text
		if plain {
			b.WriteString(t)
			continue
		}

		switch chunk.Type {
		case "link":
			b.WriteByte('[')
			b.WriteString(t)
			b.WriteByte(']')
		case "bold":
			b.WriteByte('*')
			b.WriteString(t)
			b.WriteByte('*')
		case "italic":
			b.WriteByte('_')
			b.WriteString(t)
			b.WriteByte('_')
		case "code", "pre":
			b.WriteByte('`')
			b.WriteString(t)
			b.WriteByte('`')
		case "spoiler":
			b.WriteString("||")
			b.WriteString(t)
			b.WriteString("||")
		default:
			b.WriteString(t)
		}
	}

	return strings.TrimSpace(b.String())
}

func extractMedia(m *telegram.Message) []string {
	var parts []string

	if len(m.Photo) > 0 && string(m.Photo) != "null" && string(m.Photo) != `""` {
		count := m.PhotoCount
		if count <= 1 {
			count = 1
		}
		if count > 1 {
			parts = append(parts, fmt.Sprintf("📷 %d photos", count))
		} else {
			parts = append(parts, "📷 photo")
		}
	} else {
		duration, hasDuration := parseDuration(m.DurationSeconds)

		switch m.MediaType {
		case "photo":
			parts = append(parts, "📷 photo")
		case "sticker":
			s := "sticker"
			if m.StickerEmoji != "" {
				s += " " + m.StickerEmoji
			}
			parts = append(parts, s)
		case "video_file":
			s := "🎬 video"
			if hasDuration {
				s += " (" + formatDuration(duration) + ")"
			}
			parts = append(parts, s)
		case "video_message":
			s := "📹 round video"
			if hasDuration {
				s += " (" + formatDuration(duration) + ")"
			}
			parts = append(parts, s)
		case "voice_message":
			s := "🎤 voice message"
			if hasDuration {
				s += " (" + formatDuration(duration) + ")"
			}
			parts = append(parts, s)
		case "audio_file":
			s := "🎵 audio"
			if m.Performer != "" && m.Title != "" {
				s += " — " + m.Performer + " – " + m.Title
			} else if m.Title != "" {
				s += " — " + m.Title
			}
			if hasDuration {
				s += " (" + formatDuration(duration) + ")"
			}
			parts = append(parts, s)
		case "document":
			s := "📎 file"
			if m.FileName != "" {
				s += ": " + m.FileName
			}
			parts = append(parts, s)
		case "animation":
			parts = append(parts, "🖼 GIF")
		case "live_location":
			parts = append(parts, "📍 live location")
		default:
			if m.MediaType != "" {
				parts = append(parts, "["+m.MediaType+"]")
			}
		}
	}

	if m.Poll != nil {
		question := m.Poll.Question
		if question == "" {
			question = "poll"
		}
		parts = append(parts, "📊 poll: \""+question+"\"")
	}

	if m.Contact != nil {
		name := strings.TrimSpace(m.Contact.FirstName + " " + m.Contact.LastName)
		if name == "" {
			name = "?"
		}
		parts = append(parts, "👤 contact: "+name)
	}

	if len(m.LocationInfo) > 0 && string(m.LocationInfo) != "null" {
		parts = append(parts, "📍 location")
	}

	return parts
}

func extractReactions(m *telegram.Message) string {
	if len(m.Reactions) == 0 {
		return ""
	}

	var parts []string
	for _, r := range m.Reactions {
		emoji := r.Emoji
		if emoji == "" {
			emoji = r.Type
		}
		if emoji == "" {
			emoji = "?"
		}

		count := r.Count
		if count == 0 {
			count = len(r.Recent)
		}

		switch {
		case len(r.Recent) > 0:
			names := make([]string, 0, len(r.Recent))
			for _, p := range r.Recent {
				name := p.From
				if name == "" {
					name = p.FromID
				}
				if name == "" {
					name = "?"
				}
				names = append(names, name)
			}

			extra := count - len(r.Recent)
			if extra > 0 {
				names = append(names, fmt.Sprintf("+%d", extra))
			}
			parts = append(parts, emoji+" by "+strings.Join(names, ", "))
		case count > 1:
			parts = append(parts, fmt.Sprintf("%s ×%d", emoji, count))
		default:
			parts = append(parts, emoji)
		}
	}

	return "[Reactions: " + strings.Join(parts, ", ") + "]"
}

func formatService(m *telegram.Message) (string, bool) {
	switch m.Action {
	case "phone_call":
		duration, hasDuration := parseDuration(m.DurationSeconds)
		if hasDuration {
			return "📞 phone call (" + formatDuration(duration) + ")", true
		}
		return "📞 phone call", true
	case "create_group":
		return "created the group", true
	case "edit_group_title":
		return "renamed the chat to \"" + m.Title + "\"", true
	case "edit_group_photo":
		return "changed the group photo", true
	case "delete_group_photo":
		return "removed the group photo", true
	case "invite_members":
		return "added " + strings.Join(m.Members, ", "), true
	case "remove_members":
		return "removed " + strings.Join(m.Members, ", "), true
	case "join_group_by_link":
		return "joined via invite link", true
	case "left_group":
		return "left the chat", true
	case "pin_message":
		return "📌 pinned a message", true
	case "migrate_to_supergroup":
		return "migrated to a supergroup", true
	case "score_in_game":
		score := "?"
		if len(m.Score) > 0 {
			score = strings.Trim(string(m.Score), `"`)
		}
		return "🎮 game score: " + score, true
	case "channel_chat_created":
		return "created the channel", true
	case "group_call":
		return "📞 group call", true
	default:
		if m.Action != "" {
			return "[" + m.Action + "]", true
		}
		return "", false
	}
}

func Message(m *telegram.Message, options Options) (string, bool) {
	formatted, ok := Format(m, options)
	if !ok {
		return "", false
	}
	return formatted.Line, true
}

func Format(m *telegram.Message, options Options) (FormattedMessage, bool) {
	date := formatDate(m.Date)
	sender := formatSender(m, options)

	idTag := ""
	if m.ID != 0 && !options.SkipID {
		idTag = fmt.Sprintf("#%d", m.ID)
	}

	if m.Type == "service" {
		action, ok := formatService(m)
		if !ok {
			return FormattedMessage{}, false
		}
		line := joinPrefix(date, idTag, options) + ":: " + sender + ": " + action
		return FormattedMessage{
			Line:   line,
			Sender: sender,
			Body:   action,
		}, true
	}

	text := extractText(m.Text, options.SkipEntities)
	media := []string(nil)
	if !options.SkipMedia {
		media = extractMedia(m)
	}
	reactions := ""
	if !options.SkipReactions {
		reactions = extractReactions(m)
	}

	var bodyParts []string
	if m.ForwardedFrom != "" && !options.SkipForwards {
		bodyParts = append(bodyParts, "↳ forwarded from "+m.ForwardedFrom)
	}
	if m.ReplyToMessageID != 0 && !options.SkipReplies {
		bodyParts = append(bodyParts, fmt.Sprintf("↩ reply to #%d", m.ReplyToMessageID))
	}
	if text != "" {
		bodyParts = append(bodyParts, text)
	}
	for _, med := range media {
		bodyParts = append(bodyParts, "["+med+"]")
	}

	body := strings.Join(bodyParts, "  ")
	if body == "" {
		if options.PlainDialogue {
			return FormattedMessage{}, false
		}
		body = "(empty)"
	}

	line := joinPrefix(date, idTag, options) + sender + ": " + body
	if reactions != "" {
		line += "  " + reactions
	}

	return FormattedMessage{
		Line:      line,
		Sender:    sender,
		Body:      body,
		Mergeable: text != "",
	}, true
}

func formatSender(m *telegram.Message, options Options) string {
	sender := m.From
	if sender == "" {
		sender = m.Actor
	}
	if sender == "" {
		sender = "?"
	}

	if options.ChatType != "personal_chat" {
		return sender
	}

	if isPeer(m, options) {
		if options.AnonPeer != "" {
			return options.AnonPeer
		}
		return sender
	}
	if options.AnonSelf != "" {
		return options.AnonSelf
	}
	return sender
}

func isPeer(m *telegram.Message, options Options) bool {
	if options.ChatName != "" && m.From == options.ChatName {
		return true
	}
	if options.ChatID != 0 && m.FromID == fmt.Sprintf("user%d", options.ChatID) {
		return true
	}
	return false
}

func joinPrefix(date, idTag string, options Options) string {
	var parts []string
	if !options.SkipTime {
		parts = append(parts, "["+date+"]")
	}
	if idTag != "" {
		parts = append(parts, idTag)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ") + " "
}
