package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/vo0ov/tg2txt/internal/converter"
	"github.com/vo0ov/tg2txt/internal/version"
)

const usage = `✨ tg2txt — Telegram JSON export → polished TXT

Usage:
  tg2txt [flags]

Flags:
  -h, -H        show this help message
  -i FILE        input Telegram JSON export (default: result.json)
  -o FILE        output TXT file           (default: chat.txt)
  --no-service   skip service events
  --no-header    skip the "# Chat Name" header
  -v, -V, --version
                 show build information

Examples:
  tg2txt
  tg2txt -i result.json -o chat.txt
  tg2txt -i backup.json --no-service
  tg2txt --no-header
`

func Run(args []string, stdout, stderr io.Writer) int {
	var input string
	var output string
	var noService bool
	var noHeader bool
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
	fs.StringVar(&output, "o", "chat.txt", "output TXT file")
	fs.BoolVar(&noService, "no-service", false, "skip service events")
	fs.BoolVar(&noHeader, "no-header", false, "skip the chat header")
	fs.BoolVar(&showVersion, "version", false, "show build information")
	fs.Usage = func() {}

	if err := fs.Parse(args); err != nil {
		writef(stderr, "❌ %s\n", err)
		return 2
	}

	if showVersion {
		writef(stdout, "tg2txt %s\n", version.String())
		return 0
	}

	result, err := converter.Convert(converter.Options{
		Source:      input,
		Destination: output,
		SkipService: noService,
		SkipHeader:  noHeader,
	})
	if err != nil {
		writef(stderr, "❌ %s\n", err)
		return 1
	}

	writef(stdout, "✨ Converted %d / %d messages → %s\n", result.Written, result.Total, result.Destination)
	if result.Skipped > 0 {
		writef(stdout, "   Skipped: %d\n", result.Skipped)
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
