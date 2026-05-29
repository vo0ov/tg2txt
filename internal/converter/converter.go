package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vo0ov/tg2txt/internal/formatter"
	"github.com/vo0ov/tg2txt/internal/telegram"
)

type Options struct {
	Source      string
	Destination string
	SkipService bool
	SkipHeader  bool
}

type Result struct {
	Total       int
	Written     int
	Skipped     int
	Destination string
}

func Convert(options Options) (result Result, err error) {
	result = Result{Destination: options.Destination}
	f, err := os.Open(options.Source)
	if err != nil {
		return result, fmt.Errorf("failed to open %s: %w", options.Source, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", options.Source, closeErr)
		}
	}()

	var export telegram.Export
	err = json.NewDecoder(f).Decode(&export)
	if err != nil {
		return result, fmt.Errorf("failed to parse Telegram JSON export: %w", err)
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

	var lines []string
	if export.Name != "" && !options.SkipHeader {
		lines = append(lines, "# "+export.Name, "")
	}

	result.Total = len(export.Messages)
	for i := range export.Messages {
		msg := &export.Messages[i]

		if options.SkipService && msg.Type == "service" {
			result.Skipped++
			continue
		}

		line, ok := formatter.Message(msg)
		if !ok {
			result.Skipped++
			continue
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n") + "\n"
	if _, err = out.WriteString(content); err != nil {
		return result, fmt.Errorf("failed to write output file: %w", err)
	}

	result.Written = result.Total - result.Skipped
	return result, nil
}
