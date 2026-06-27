package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPlainDialoguePresetWithAnonFlags(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")

	export := `{
		"name": "Боб",
		"type": "personal_chat",
		"id": 1828671611,
		"messages": [
			{
				"id": 292386,
				"type": "message",
				"date": "2025-09-23T20:51:31",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "тест"
			},
			{
				"id": 292387,
				"type": "message",
				"date": "2025-09-23T20:52:31",
				"from": "Боб",
				"from_id": "user1828671611",
				"forwarded_from": "Carol",
				"reply_to_message_id": 292386,
				"media_type": "photo",
				"text": [
					{"type": "bold", "text": "Привет"}
				],
				"reactions": [
					{"emoji": "👍", "count": 1}
				]
			},
			{
				"id": 292388,
				"type": "service",
				"date": "2025-09-23T20:53:31",
				"actor": "Боб",
				"action": "pin_message"
			}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--plain-dialogue",
		"--anon-peer", "Боб",
		"--anon-self", "Алекс",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "Алекс: тест\nБоб: Привет\n"
	if string(got) != want {
		t.Fatalf("output = %q, want %q", string(got), want)
	}
}

func TestJoinMessagesMergesNearbyMessagesFromSameSender(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")

	export := `{
		"name": "Боб",
		"type": "personal_chat",
		"id": 1828671611,
		"messages": [
			{
				"id": 1,
				"type": "message",
				"date": "2025-09-23T20:51:31",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "го"
			},
			{
				"id": 2,
				"type": "message",
				"date": "2025-09-23T20:51:36",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "а где?"
			},
			{
				"id": 3,
				"type": "message",
				"date": "2025-09-23T20:51:40",
				"from": "Боб",
				"from_id": "user1828671611",
				"text": "тут"
			},
			{
				"id": 4,
				"type": "message",
				"date": "2025-09-23T20:51:45",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "и давай чуть позже"
			},
			{
				"id": 5,
				"type": "message",
				"date": "2025-09-23T20:52:05",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "уже поздно"
			}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--plain-dialogue",
		"--join-messages",
		"--anon-peer", "Боб",
		"--anon-self", "Алекс",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "Алекс: го\nа где?\nБоб: тут\nАлекс: и давай чуть позже\nАлекс: уже поздно\n"
	if string(got) != want {
		t.Fatalf("output = %q, want %q", string(got), want)
	}
}

func TestJoinMessagesUsesCustomSeparator(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")

	export := `{
		"name": "Боб",
		"type": "personal_chat",
		"id": 1828671611,
		"messages": [
			{
				"id": 1,
				"type": "message",
				"date": "2025-09-23T20:51:31",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "го"
			},
			{
				"id": 2,
				"type": "message",
				"date": "2025-09-23T20:51:36",
				"from": "алекс",
				"from_id": "user1770663897",
				"text": "а где?"
			}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--plain-dialogue",
		"--join-messages",
		"--join-separator", " / ",
		"--anon-self", "Алекс",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "Алекс: го / а где?\n"
	if string(got) != want {
		t.Fatalf("output = %q, want %q", string(got), want)
	}
}

func TestJoinWindowMustBeNonNegative(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--join-messages", "--join-window", "-1"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if stderr.String() != "❌ --join-window must be non-negative\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestActivityPNGGeneratesChartFile(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")
	activity := filepath.Join(dir, "activity.png")

	export := `{
		"name": "Team Chat",
		"type": "private_group",
		"id": 1,
		"messages": [
			{"id": 1, "type": "message", "date": "2025-09-23T20:51:31", "from": "Alice", "text": "one"},
			{"id": 2, "type": "service", "date": "2025-09-24T20:51:31", "actor": "Alice", "action": "edit_group_photo"},
			{"id": 3, "type": "message", "date": "2025-09-25T20:51:31", "from": "Bob", "text": "two"}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--activity-png", activity,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	info, err := os.Stat(activity)
	if err != nil {
		t.Fatalf("os.Stat(activity) error = %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("activity.png is empty")
	}

	if !bytes.Contains(stdout.Bytes(), []byte("📈 Activity chart → "+activity)) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestActivityPNGUsesDefaultNameWhenValueIsOmitted(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")
	defaultChart := filepath.Join(dir, "chart.png")

	export := `{
		"name": "Team Chat",
		"type": "private_group",
		"id": 1,
		"messages": [
			{"id": 1, "type": "message", "date": "2025-09-23T20:51:31", "from": "Alice", "text": "one"}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	chdirErr := os.Chdir(dir)
	if chdirErr != nil {
		t.Fatalf("os.Chdir() error = %v", chdirErr)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--activity-png",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	info, err := os.Stat(defaultChart)
	if err != nil {
		t.Fatalf("os.Stat(chart.png) error = %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("chart.png is empty")
	}

	if !bytes.Contains(stdout.Bytes(), []byte("📈 Activity chart → chart.png")) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestStatsUsesDefaultNameWhenValueIsOmitted(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "result.json")
	output := filepath.Join(dir, "chat.txt")
	defaultStats := filepath.Join(dir, "stats.txt")

	export := `{
		"name": "Team Chat",
		"type": "private_group",
		"id": 1,
		"messages": [
			{"id": 1, "type": "message", "date": "2025-09-23T20:51:31", "from": "Alice", "text": "one"}
		]
	}`

	if err := os.WriteFile(input, []byte(export), 0600); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	chdirErr := os.Chdir(dir)
	if chdirErr != nil {
		t.Fatalf("os.Chdir() error = %v", chdirErr)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"--input", input,
		"--output", output,
		"--stats",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}

	got, err := os.ReadFile(defaultStats)
	if err != nil {
		t.Fatalf("os.ReadFile(stats.txt) error = %v", err)
	}
	if !bytes.Contains(got, []byte("Total messages: 1")) {
		t.Fatalf("stats.txt = %q", string(got))
	}

	if !bytes.Contains(stdout.Bytes(), []byte("📊 Stats report → stats.txt")) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
