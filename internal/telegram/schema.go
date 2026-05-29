package telegram

import "encoding/json"

type Export struct {
	Name     string    `json:"name"`
	Messages []Message `json:"messages"`
}

type Message struct {
	ID               int64           `json:"id"`
	Type             string          `json:"type"`
	Date             string          `json:"date"`
	From             string          `json:"from"`
	Actor            string          `json:"actor"`
	Text             json.RawMessage `json:"text"` // string | []TextChunk
	MediaType        string          `json:"media_type"`
	Photo            json.RawMessage `json:"photo"` // string | absent
	PhotoCount       int             `json:"photo_count"`
	StickerEmoji     string          `json:"sticker_emoji"`
	DurationSeconds  json.RawMessage `json:"duration_seconds"` // int | string
	Title            string          `json:"title"`
	Performer        string          `json:"performer"`
	FileName         string          `json:"file_name"`
	ForwardedFrom    string          `json:"forwarded_from"`
	ReplyToMessageID int64           `json:"reply_to_message_id"`
	Reactions        []Reaction      `json:"reactions"`
	Action           string          `json:"action"`
	Members          []string        `json:"members"`
	Poll             *Poll           `json:"poll"`
	Contact          *Contact        `json:"contact"`
	LocationInfo     json.RawMessage `json:"location_information"`
	Score            json.RawMessage `json:"score"`
}

type TextChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Reaction struct {
	Type   string         `json:"type"`
	Emoji  string         `json:"emoji"`
	Count  int            `json:"count"`
	Recent []ReactionFrom `json:"recent"`
}

type ReactionFrom struct {
	From   string `json:"from"`
	FromID string `json:"from_id"`
}

type Poll struct {
	Question string `json:"question"`
}

type Contact struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
