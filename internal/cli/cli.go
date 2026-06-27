package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/vo0ov/tg2txt/internal/converter"
	"github.com/vo0ov/tg2txt/internal/formatter"
	"github.com/vo0ov/tg2txt/internal/version"
)

const usage = `✨ tg2txt — Telegram JSON export → polished TXT

Usage:
  tg2txt [flags]

Flags:
  -h, -H        show this help message
  -i, --input FILE
                 input Telegram JSON export (default: result.json)
  -o, --output FILE
                 output TXT file           (default: chat.txt)
  --activity-png FILE
                 output PNG activity chart by day (default: chart.png)
  --stats FILE
                 output TXT chat statistics report (default: stats.txt)
  --no-header    skip the "# Chat Name" header
  --no-time      skip message timestamps
  --no-id        skip Telegram message ids
  --no-service   skip service events
  --no-media     skip media/contact/location/poll markers
  --no-reactions skip reaction summaries
  --no-entities  skip Telegram entity formatting
  --no-forwards  skip forwarded-from context
  --no-replies   skip reply context
  --plain-dialogue
                 preset for "Name: text" dialogue output
  --anon-peer NAME
                 rename the peer in personal_chat exports
  --anon-self NAME
                 rename the export owner in personal_chat exports
  --join-messages
                 merge nearby consecutive messages from the same sender
  --join-separator TEXT
                 separator for merged message bodies (default: \n)
  --join-window SECONDS
                 merge messages within this many seconds (default: 15)
  -v, -V, --version
                 show build information

Examples:
  tg2txt
  tg2txt -i result.json -o chat.txt
  tg2txt -i result.json -o chat.txt --activity-png activity.png
  tg2txt -i backup.json --no-service
  tg2txt --plain-dialogue --anon-peer Bob --anon-self Alex
`

func Run(args []string, stdout, stderr io.Writer) int {
	var input string
	var output string
	var activityPNG string
	var statsTXT string
	var noService bool
	var noHeader bool
	var noTime bool
	var noID bool
	var noMedia bool
	var noReactions bool
	var noEntities bool
	var noForwards bool
	var noReplies bool
	var plainDialogue bool
	var anonPeer string
	var anonSelf string
	var joinMessages bool
	var joinSeparator string
	var joinWindow int
	var showVersion bool

	if wantsHelp(args) {
		write(stdout, usage)
		return 0
	}
	if wantsVersion(args) {
		writef(stdout, "tg2txt %s\n", version.String())
		return 0
	}

	fs := flag.NewFlagSet("tg2txt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&input, "i", "result.json", "input Telegram JSON export")
	fs.StringVar(&input, "input", "result.json", "input Telegram JSON export")
	fs.StringVar(&output, "o", "chat.txt", "output TXT file")
	fs.StringVar(&output, "output", "chat.txt", "output TXT file")
	fs.StringVar(&activityPNG, "activity-png", "", "output PNG activity chart by day")
	fs.StringVar(&statsTXT, "stats", "", "output TXT chat statistics report")
	fs.BoolVar(&noService, "no-service", false, "skip service events")
	fs.BoolVar(&noHeader, "no-header", false, "skip the chat header")
	fs.BoolVar(&noTime, "no-time", false, "skip message timestamps")
	fs.BoolVar(&noID, "no-id", false, "skip Telegram message ids")
	fs.BoolVar(&noMedia, "no-media", false, "skip media/contact/location/poll markers")
	fs.BoolVar(&noReactions, "no-reactions", false, "skip reaction summaries")
	fs.BoolVar(&noEntities, "no-entities", false, "skip Telegram entity formatting")
	fs.BoolVar(&noForwards, "no-forwards", false, "skip forwarded-from context")
	fs.BoolVar(&noReplies, "no-replies", false, "skip reply context")
	fs.BoolVar(&plainDialogue, "plain-dialogue", false, "preset for dialogue-only output")
	fs.StringVar(&anonPeer, "anon-peer", "", "rename the peer in personal_chat exports")
	fs.StringVar(&anonSelf, "anon-self", "", "rename the export owner in personal_chat exports")
	fs.BoolVar(&joinMessages, "join-messages", false, "merge nearby consecutive messages from the same sender")
	fs.StringVar(&joinSeparator, "join-separator", `\n`, "separator for merged message bodies")
	fs.IntVar(&joinWindow, "join-window", 15, "merge messages within this many seconds")
	fs.BoolVar(&showVersion, "version", false, "show build information")
	fs.Usage = func() {}

	if err := fs.Parse(normalizeArgs(args)); err != nil {
		writef(stderr, "❌ %s\n", err)
		return 2
	}

	if showVersion {
		writef(stdout, "tg2txt %s\n", version.String())
		return 0
	}
	if joinWindow < 0 {
		writef(stderr, "❌ --join-window must be non-negative\n")
		return 2
	}

	if plainDialogue {
		noHeader = true
		noTime = true
		noID = true
		noService = true
		noMedia = true
		noReactions = true
		noEntities = true
		noForwards = true
		noReplies = true
	}

	result, err := converter.Convert(converter.Options{
		Source:              input,
		Destination:         output,
		ActivityDestination: activityPNG,
		StatsDestination:    statsTXT,
		SkipService:         noService,
		SkipHeader:          noHeader,
		Join: converter.JoinOptions{
			Enabled:       joinMessages,
			Separator:     decodeSeparator(joinSeparator),
			WindowSeconds: joinWindow,
		},
		Format: formatter.Options{
			SkipTime:      noTime,
			SkipID:        noID,
			SkipMedia:     noMedia,
			SkipReactions: noReactions,
			SkipEntities:  noEntities,
			SkipForwards:  noForwards,
			SkipReplies:   noReplies,
			PlainDialogue: plainDialogue,
			AnonPeer:      anonPeer,
			AnonSelf:      anonSelf,
		},
	})
	if err != nil {
		writef(stderr, "❌ %s\n", err)
		return 1
	}

	writef(stdout, "✨ Converted %d / %d messages → %s\n", result.Written, result.Total, result.Destination)
	if result.Skipped > 0 {
		writef(stdout, "   Skipped: %d\n", result.Skipped)
	}
	if result.ActivityDestination != "" {
		writef(stdout, "📈 Activity chart → %s\n", result.ActivityDestination)
	}
	if result.StatsDestination != "" {
		writef(stdout, "📊 Stats report → %s\n", result.StatsDestination)
	}
	return 0
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-h", "-H", "--help":
			return true
		}
	}
	return false
}

func wantsVersion(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-v", "-V", "--version":
			return true
		}
	}
	return false
}

func write(w io.Writer, s string) {
	_, _ = fmt.Fprint(w, s)
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func decodeSeparator(raw string) string {
	replacer := strings.NewReplacer(
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
	)
	return replacer.Replace(raw)
}

func normalizeArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--activity-png":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				normalized = append(normalized, arg, defaultActivityPNGPath())
				continue
			}
		case "--stats":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				normalized = append(normalized, arg, defaultStatsTXTPath())
				continue
			}
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

func defaultActivityPNGPath() string {
	return filepath.Clean("chart.png")
}

func defaultStatsTXTPath() string {
	return filepath.Clean("stats.txt")
}
