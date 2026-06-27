package stats

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vo0ov/tg2txt/internal/telegram"
)

const (
	telegramTimeLayout = "2006-01-02T15:04:05"
	responseWindow     = 12 * time.Hour
	fastResponse       = 5 * time.Minute
	normalResponse     = time.Hour
)

type Options struct {
	Destination string
}

func WriteTXT(export telegram.Export, options Options) (err error) {
	report := Render(export)

	mkdirErr := os.MkdirAll(filepath.Dir(options.Destination), 0750)
	if mkdirErr != nil {
		return fmt.Errorf("failed to create stats directory: %w", mkdirErr)
	}

	out, err := os.Create(options.Destination)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", options.Destination, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", options.Destination, closeErr)
		}
	}()

	_, writeErr := out.WriteString(report)
	if writeErr != nil {
		return fmt.Errorf("failed to write stats file: %w", writeErr)
	}
	return nil
}

func Render(export telegram.Export) string {
	report := analyze(export)
	var b strings.Builder

	writeLine(&b, "tg2txt chat stats")
	if export.Name != "" {
		writeLine(&b, "Chat: %s", export.Name)
	}
	writeLine(&b, "Response window: %s", formatDuration(responseWindow))
	writeLine(&b, "")

	writeLine(&b, "Basic")
	writeLine(&b, "  Total messages: %d", report.TotalMessages)
	writeLine(&b, "  Participants: %d", len(report.Participants))
	writeLine(&b, "  First day: %s", formatDay(report.FirstDay))
	writeLine(&b, "  Last day: %s", formatDay(report.LastDay))
	writeLine(&b, "  Calendar days: %d", report.CalendarDays)
	writeLine(&b, "  Active days: %d", report.ActiveDays)
	writeLine(&b, "  Zero-message days: %d", report.ZeroDays)
	writeLine(&b, "  Avg messages/day: %.2f", report.AvgMessagesPerDay)
	writeLine(&b, "  Avg messages/active day: %.2f", report.AvgMessagesPerActiveDay)
	writeLine(&b, "")

	writeLine(&b, "Daily Activity")
	writeLine(&b, "  Min messages/day: %d on %s", report.MinMessagesDay, formatDay(report.MinMessagesDayDate))
	writeLine(&b, "  Min messages/active day: %d on %s", report.MinActiveMessagesDay, formatDay(report.MinActiveMessagesDayDate))
	writeLine(&b, "  Max messages/day: %d on %s", report.MaxMessagesDay, formatDay(report.MaxMessagesDayDate))
	writeLine(&b, "  Longest active streak: %d days", report.LongestActiveStreak)
	writeLine(&b, "  Longest gap between active days: %d days", report.LongestInactiveGap)
	writeLine(&b, "  Most active hour: %s", formatHour(report.MostActiveHour))
	writeLine(&b, "  Weekday distribution: %s", formatWeekdays(report.WeekdayCounts))
	writeLine(&b, "  Night messages (00:00-05:59): %d (%.2f%%)", report.NightMessages, percent(report.NightMessages, report.TotalMessages))
	writeLine(&b, "")

	writeLine(&b, "Participation And Balance")
	writeLine(&b, "  Most messages: %s", formatParticipantTotal(report.MostMessages))
	writeLine(&b, "  Fewest messages: %s", formatParticipantTotal(report.FewestMessages))
	writeLine(&b, "  Dominance coefficient: %.2f%%", report.DominanceCoefficient)
	writeLine(&b, "  Evenness index: %.3f", report.EvennessIndex)
	writeLine(&b, "")

	writeLine(&b, "Content Mix")
	writeLine(&b, "  Text messages: %d (%.2f%%)", report.TextMessages, percent(report.TextMessages, report.TotalMessages))
	writeLine(&b, "  Media messages: %d (%.2f%%)", report.MediaMessages, percent(report.MediaMessages, report.TotalMessages))
	writeLine(&b, "  Service events: %d", report.ServiceEvents)
	writeLine(&b, "  Replies: %d", report.Replies)
	writeLine(&b, "  Forwards: %d", report.Forwards)
	writeLine(&b, "")

	writeLine(&b, "Message Type Top")
	if len(report.MessageTypes) == 0 {
		writeLine(&b, "  n/a")
	} else {
		for _, item := range report.MessageTypes {
			writeLine(&b, "  %s: %d (%.2f%%)", item.Name, item.Count, percent(item.Count, report.MessageTypeTotal))
		}
	}
	writeLine(&b, "")

	writeLine(&b, "Participants")
	for _, p := range report.Participants {
		writeLine(&b, "  %s", p.Name)
		writeLine(&b, "    Messages: %d (%.2f%%)", p.Messages, p.Percent)
		writeLine(&b, "    Active days: %d", p.ActiveDays)
		writeLine(&b, "    Daily min active / min any / max / avg active: %d / %d / %d / %.2f", p.MinActiveDayMessages, p.MinAnyDayMessages, p.MaxDayMessages, p.AvgActiveDayMessages)
		writeLine(&b, "    Days as top participant: %d", p.TopDays)
		writeLine(&b, "    Days absent while chat was active: %d", p.AbsentActiveChatDays)
		writeLine(&b, "    Median response: %s", formatDurationOrNA(p.MedianResponse, p.ValidResponses))
		writeLine(&b, "    P75 response: %s", formatDurationOrNA(p.P75Response, p.ValidResponses))
		writeLine(&b, "    Valid responses: %d", p.ValidResponses)
		writeLine(&b, "    Fast / normal / long responses: %d / %d / %d", p.FastResponses, p.NormalResponses, p.LongResponses)
		writeLine(&b, "    Fast response share: %.2f%%", percent(p.FastResponses, p.ValidResponses))
		writeLine(&b, "    Avg previous series length before response: %.2f", p.AvgSeriesBeforeResponse)
		writeLine(&b, "    Consecutive series count / avg / max: %d / %.2f / %d", p.SeriesCount, p.AvgSeriesLength, p.MaxSeriesLength)
		writeLine(&b, "    Unanswered messages within %s: %d", formatDuration(responseWindow), p.UnansweredMessages)
		writeLine(&b, "    Avg message length chars / words: %.2f / %.2f", p.AvgChars, p.AvgWords)
		writeLine(&b, "    Media / replies / forwards: %d / %d / %d", p.MediaMessages, p.Replies, p.Forwards)
		writeLine(&b, "    Reaction top:")
		if len(p.ReactionTop) == 0 {
			writeLine(&b, "      n/a")
		} else {
			for _, item := range p.ReactionTop {
				writeLine(&b, "      %s: %d", item.Name, item.Count)
			}
		}
		writeLine(&b, "    Day starts / day ends: %d / %d", p.DayStarts, p.DayEnds)
		writeLine(&b, "    Most active hour: %s", formatHour(p.MostActiveHour))
	}

	return b.String()
}

type report struct {
	TotalMessages            int
	ServiceEvents            int
	TextMessages             int
	MediaMessages            int
	Replies                  int
	Forwards                 int
	FirstDay                 time.Time
	LastDay                  time.Time
	CalendarDays             int
	ActiveDays               int
	ZeroDays                 int
	AvgMessagesPerDay        float64
	AvgMessagesPerActiveDay  float64
	MinMessagesDay           int
	MinMessagesDayDate       time.Time
	MinActiveMessagesDay     int
	MinActiveMessagesDayDate time.Time
	MaxMessagesDay           int
	MaxMessagesDayDate       time.Time
	LongestActiveStreak      int
	LongestInactiveGap       int
	MostActiveHour           int
	WeekdayCounts            [7]int
	NightMessages            int
	MostMessages             participantTotal
	FewestMessages           participantTotal
	DominanceCoefficient     float64
	EvennessIndex            float64
	MessageTypeTotal         int
	MessageTypes             []typeCount
	Participants             []participantReport
}

type typeCount struct {
	Name  string
	Count int
}

type participantTotal struct {
	Name  string
	Count int
}

type participantReport struct {
	Name                    string
	Messages                int
	Percent                 float64
	ActiveDays              int
	MinActiveDayMessages    int
	MinAnyDayMessages       int
	MaxDayMessages          int
	AvgActiveDayMessages    float64
	TopDays                 int
	AbsentActiveChatDays    int
	MedianResponse          time.Duration
	P75Response             time.Duration
	ValidResponses          int
	FastResponses           int
	NormalResponses         int
	LongResponses           int
	AvgSeriesBeforeResponse float64
	SeriesCount             int
	AvgSeriesLength         float64
	MaxSeriesLength         int
	UnansweredMessages      int
	AvgChars                float64
	AvgWords                float64
	MediaMessages           int
	Replies                 int
	Forwards                int
	ReactionTop             []typeCount
	DayStarts               int
	DayEnds                 int
	MostActiveHour          int
}

type participantAccumulator struct {
	name                    string
	messages                int
	daily                   map[time.Time]int
	chars                   int
	words                   int
	media                   int
	replies                 int
	forwards                int
	reactions               map[string]int
	hours                   [24]int
	dayStarts               int
	dayEnds                 int
	topDays                 int
	responses               []time.Duration
	fastResponses           int
	normalResponses         int
	longResponses           int
	seriesBeforeResponseSum int
	seriesCount             int
	seriesLengthSum         int
	maxSeriesLength         int
	unansweredMessages      int
}

type messageEvent struct {
	sender    string
	when      time.Time
	day       time.Time
	text      string
	chars     int
	words     int
	hasMedia  bool
	hasReply  bool
	hasFwd    bool
	reactions []typeCount
}

type turn struct {
	sender string
	start  time.Time
	end    time.Time
	count  int
}

func analyze(export telegram.Export) report {
	events, serviceEvents := collectEvents(export.Messages)
	out := report{
		TotalMessages:    len(events),
		ServiceEvents:    serviceEvents,
		MessageTypes:     collectTypeCounts(export.Messages),
		MessageTypeTotal: len(export.Messages),
	}
	if len(events) == 0 {
		return out
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].when.Before(events[j].when)
	})

	participants := make(map[string]*participantAccumulator)
	dayCounts := make(map[time.Time]int)
	dayEvents := make(map[time.Time][]messageEvent)
	hourCounts := [24]int{}

	out.FirstDay = events[0].day
	out.LastDay = events[len(events)-1].day
	for _, event := range events {
		acc := ensureParticipant(participants, event.sender)
		acc.messages++
		acc.chars += event.chars
		acc.words += event.words
		acc.daily[event.day]++
		acc.hours[event.when.Hour()]++

		if event.text != "" {
			out.TextMessages++
		}
		if event.hasMedia {
			out.MediaMessages++
			acc.media++
		}
		if event.hasReply {
			out.Replies++
			acc.replies++
		}
		if event.hasFwd {
			out.Forwards++
			acc.forwards++
		}
		for _, reaction := range event.reactions {
			acc.reactions[reaction.Name] += reaction.Count
		}
		if event.when.Hour() < 6 {
			out.NightMessages++
		}

		dayCounts[event.day]++
		dayEvents[event.day] = append(dayEvents[event.day], event)
		hourCounts[event.when.Hour()]++
		out.WeekdayCounts[int(event.when.Weekday())]++
	}

	out.CalendarDays = inclusiveDays(out.FirstDay, out.LastDay)
	out.ActiveDays = len(dayCounts)
	out.ZeroDays = out.CalendarDays - out.ActiveDays
	out.AvgMessagesPerDay = safeDivFloat(out.TotalMessages, out.CalendarDays)
	out.AvgMessagesPerActiveDay = safeDivFloat(out.TotalMessages, out.ActiveDays)
	out.MostActiveHour = maxIndex(hourCounts[:])

	fillDailyStats(&out, dayCounts)
	fillParticipantDayStats(participants, dayCounts)
	fillDayBoundaryStats(participants, dayEvents)
	fillTurns(participants, events)
	fillBalanceStats(&out, participants)

	out.Participants = buildParticipantReports(participants, out.TotalMessages, out.ActiveDays, out.CalendarDays)
	return out
}

func collectEvents(messages []telegram.Message) ([]messageEvent, int) {
	events := make([]messageEvent, 0, len(messages))
	serviceEvents := 0
	for _, msg := range messages {
		if msg.Type == "service" {
			serviceEvents++
			continue
		}
		if msg.Type != "message" {
			continue
		}

		when, err := time.Parse(telegramTimeLayout, msg.Date)
		if err != nil {
			continue
		}

		text := extractPlainText(msg.Text)
		events = append(events, messageEvent{
			sender:    senderName(msg),
			when:      when,
			day:       dayStart(when),
			text:      text,
			chars:     len([]rune(text)),
			words:     len(strings.Fields(text)),
			hasMedia:  hasMedia(msg),
			hasReply:  msg.ReplyToMessageID != 0,
			hasFwd:    msg.ForwardedFrom != "",
			reactions: messageReactions(msg),
		})
	}
	return events, serviceEvents
}

func collectTypeCounts(messages []telegram.Message) []typeCount {
	counts := make(map[string]int)
	for _, msg := range messages {
		counts[messageTypeName(msg)]++
	}

	items := make([]typeCount, 0, len(counts))
	for name, count := range counts {
		items = append(items, typeCount{Name: name, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	return items
}

func messageTypeName(msg telegram.Message) string {
	if msg.Type == "service" {
		if msg.Action != "" {
			return "service:" + msg.Action
		}
		return "service"
	}
	if msg.MediaType != "" {
		return msg.MediaType
	}
	if len(msg.Photo) > 0 && string(msg.Photo) != "null" && string(msg.Photo) != `""` {
		return "photo"
	}
	if msg.Poll != nil {
		return "poll"
	}
	if msg.Contact != nil {
		return "contact"
	}
	if len(msg.LocationInfo) > 0 && string(msg.LocationInfo) != "null" {
		return "location"
	}
	if extractPlainText(msg.Text) != "" {
		return "text"
	}
	if msg.Type != "" {
		return msg.Type
	}
	return "unknown"
}

func fillDailyStats(out *report, dayCounts map[time.Time]int) {
	first := true
	activeFirst := true
	currentStreak := 0
	lastActive := time.Time{}
	for day := out.FirstDay; !day.After(out.LastDay); day = day.AddDate(0, 0, 1) {
		count := dayCounts[day]
		if first || count < out.MinMessagesDay {
			out.MinMessagesDay = count
			out.MinMessagesDayDate = day
		}
		if first || count > out.MaxMessagesDay {
			out.MaxMessagesDay = count
			out.MaxMessagesDayDate = day
		}
		first = false

		if count > 0 {
			if activeFirst || count < out.MinActiveMessagesDay {
				out.MinActiveMessagesDay = count
				out.MinActiveMessagesDayDate = day
			}
			activeFirst = false

			currentStreak++
			if currentStreak > out.LongestActiveStreak {
				out.LongestActiveStreak = currentStreak
			}

			if !lastActive.IsZero() {
				gap := inclusiveDays(lastActive, day) - 2
				if gap > out.LongestInactiveGap {
					out.LongestInactiveGap = gap
				}
			}
			lastActive = day
			continue
		}
		currentStreak = 0
	}
}

func fillParticipantDayStats(participants map[string]*participantAccumulator, dayCounts map[time.Time]int) {
	for day := range dayCounts {
		maxCount := 0
		for _, acc := range participants {
			if acc.daily[day] > maxCount {
				maxCount = acc.daily[day]
			}
		}
		for _, acc := range participants {
			if maxCount > 0 && acc.daily[day] == maxCount {
				acc.topDays++
			}
		}
	}
}

func fillDayBoundaryStats(participants map[string]*participantAccumulator, dayEvents map[time.Time][]messageEvent) {
	for _, events := range dayEvents {
		sort.Slice(events, func(i, j int) bool {
			return events[i].when.Before(events[j].when)
		})
		ensureParticipant(participants, events[0].sender).dayStarts++
		ensureParticipant(participants, events[len(events)-1].sender).dayEnds++
	}
}

func fillTurns(participants map[string]*participantAccumulator, events []messageEvent) {
	turns := buildTurns(events)
	for i, current := range turns {
		acc := ensureParticipant(participants, current.sender)
		acc.seriesCount++
		acc.seriesLengthSum += current.count
		if current.count > acc.maxSeriesLength {
			acc.maxSeriesLength = current.count
		}

		if i > 0 {
			previous := turns[i-1]
			delta := current.start.Sub(previous.end)
			if delta > 0 && delta <= responseWindow {
				acc.responses = append(acc.responses, delta)
				acc.seriesBeforeResponseSum += previous.count
				switch {
				case delta <= fastResponse:
					acc.fastResponses++
				case delta <= normalResponse:
					acc.normalResponses++
				default:
					acc.longResponses++
				}
			}
		}

		if i == len(turns)-1 {
			acc.unansweredMessages += current.count
			continue
		}
		next := turns[i+1]
		if next.start.Sub(current.end) > responseWindow {
			acc.unansweredMessages += current.count
		}
	}
}

func fillBalanceStats(out *report, participants map[string]*participantAccumulator) {
	names := sortedParticipantNames(participants)
	if len(names) == 0 {
		return
	}

	out.MostMessages = participantTotal{Name: names[0], Count: participants[names[0]].messages}
	out.FewestMessages = out.MostMessages

	entropy := 0.0
	for _, name := range names {
		count := participants[name].messages
		if count > out.MostMessages.Count || (count == out.MostMessages.Count && name < out.MostMessages.Name) {
			out.MostMessages = participantTotal{Name: name, Count: count}
		}
		if count < out.FewestMessages.Count || (count == out.FewestMessages.Count && name < out.FewestMessages.Name) {
			out.FewestMessages = participantTotal{Name: name, Count: count}
		}

		p := float64(count) / float64(out.TotalMessages)
		if p > 0 {
			entropy -= p * math.Log(p)
		}
	}

	out.DominanceCoefficient = percent(out.MostMessages.Count, out.TotalMessages)
	if len(names) == 1 {
		out.EvennessIndex = 1
	} else {
		out.EvennessIndex = entropy / math.Log(float64(len(names)))
	}
}

func buildParticipantReports(participants map[string]*participantAccumulator, totalMessages, chatActiveDays, calendarDays int) []participantReport {
	names := sortedParticipantNames(participants)
	reports := make([]participantReport, 0, len(names))
	for _, name := range names {
		acc := participants[name]
		activeDays := len(acc.daily)
		minActive := 0
		maxDay := 0
		sumActive := 0
		first := true
		for _, count := range acc.daily {
			if first || count < minActive {
				minActive = count
			}
			if count > maxDay {
				maxDay = count
			}
			sumActive += count
			first = false
		}

		minAny := minActive
		if activeDays < calendarDays {
			minAny = 0
		}

		responses := append([]time.Duration(nil), acc.responses...)
		sort.Slice(responses, func(i, j int) bool {
			return responses[i] < responses[j]
		})

		validResponses := len(responses)
		reports = append(reports, participantReport{
			Name:                    name,
			Messages:                acc.messages,
			Percent:                 percent(acc.messages, totalMessages),
			ActiveDays:              activeDays,
			MinActiveDayMessages:    minActive,
			MinAnyDayMessages:       minAny,
			MaxDayMessages:          maxDay,
			AvgActiveDayMessages:    safeDivFloat(sumActive, activeDays),
			TopDays:                 acc.topDays,
			AbsentActiveChatDays:    chatActiveDays - activeDays,
			MedianResponse:          percentileDuration(responses, 0.50),
			P75Response:             percentileDuration(responses, 0.75),
			ValidResponses:          validResponses,
			FastResponses:           acc.fastResponses,
			NormalResponses:         acc.normalResponses,
			LongResponses:           acc.longResponses,
			AvgSeriesBeforeResponse: safeDivFloat(acc.seriesBeforeResponseSum, validResponses),
			SeriesCount:             acc.seriesCount,
			AvgSeriesLength:         safeDivFloat(acc.seriesLengthSum, acc.seriesCount),
			MaxSeriesLength:         acc.maxSeriesLength,
			UnansweredMessages:      acc.unansweredMessages,
			AvgChars:                safeDivFloat(acc.chars, acc.messages),
			AvgWords:                safeDivFloat(acc.words, acc.messages),
			MediaMessages:           acc.media,
			Replies:                 acc.replies,
			Forwards:                acc.forwards,
			ReactionTop:             typeCountsFromMap(acc.reactions),
			DayStarts:               acc.dayStarts,
			DayEnds:                 acc.dayEnds,
			MostActiveHour:          maxIndex(acc.hours[:]),
		})
	}
	return reports
}

func buildTurns(events []messageEvent) []turn {
	if len(events) == 0 {
		return nil
	}

	turns := []turn{{
		sender: events[0].sender,
		start:  events[0].when,
		end:    events[0].when,
		count:  1,
	}}
	for _, event := range events[1:] {
		last := &turns[len(turns)-1]
		if event.sender == last.sender {
			last.end = event.when
			last.count++
			continue
		}

		turns = append(turns, turn{
			sender: event.sender,
			start:  event.when,
			end:    event.when,
			count:  1,
		})
	}
	return turns
}

func ensureParticipant(participants map[string]*participantAccumulator, name string) *participantAccumulator {
	acc, ok := participants[name]
	if ok {
		return acc
	}

	acc = &participantAccumulator{
		name:      name,
		daily:     make(map[time.Time]int),
		reactions: make(map[string]int),
	}
	participants[name] = acc
	return acc
}

func messageReactions(msg telegram.Message) []typeCount {
	if len(msg.Reactions) == 0 {
		return nil
	}

	items := make([]typeCount, 0, len(msg.Reactions))
	for _, reaction := range msg.Reactions {
		name := reaction.Emoji
		if name == "" {
			name = reaction.Type
		}
		if name == "" {
			name = "unknown"
		}

		count := reaction.Count
		if count == 0 {
			count = len(reaction.Recent)
		}
		if count == 0 {
			count = 1
		}

		items = append(items, typeCount{
			Name:  name,
			Count: count,
		})
	}
	return items
}

func sortedParticipantNames(participants map[string]*participantAccumulator) []string {
	names := make([]string, 0, len(participants))
	for name := range participants {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		left := participants[names[i]]
		right := participants[names[j]]
		if left.messages == right.messages {
			return names[i] < names[j]
		}
		return left.messages > right.messages
	})
	return names
}

func typeCountsFromMap(counts map[string]int) []typeCount {
	items := make([]typeCount, 0, len(counts))
	for name, count := range counts {
		items = append(items, typeCount{
			Name:  name,
			Count: count,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	return items
}

func extractPlainText(raw json.RawMessage) string {
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
		if err := json.Unmarshal(elem, &chunk); err == nil {
			b.WriteString(chunk.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func senderName(msg telegram.Message) string {
	if msg.From != "" {
		return msg.From
	}
	if msg.FromID != "" {
		return msg.FromID
	}
	return "?"
}

func hasMedia(msg telegram.Message) bool {
	if msg.MediaType != "" || msg.Poll != nil || msg.Contact != nil {
		return true
	}
	if len(msg.LocationInfo) > 0 && string(msg.LocationInfo) != "null" {
		return true
	}
	return len(msg.Photo) > 0 && string(msg.Photo) != "null" && string(msg.Photo) != `""`
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func inclusiveDays(first, last time.Time) int {
	if first.IsZero() || last.IsZero() || last.Before(first) {
		return 0
	}
	return int(last.Sub(first).Hours()/24) + 1
}

func percentileDuration(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(math.Ceil(percentile*float64(len(values)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func maxIndex(values []int) int {
	bestIndex := 0
	bestValue := 0
	for i, value := range values {
		if i == 0 || value > bestValue {
			bestIndex = i
			bestValue = value
		}
	}
	return bestIndex
}

func safeDivFloat(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func percent(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

func writeLine(b *strings.Builder, format string, args ...any) {
	if len(args) == 0 {
		b.WriteString(format)
		b.WriteByte('\n')
		return
	}
	_, _ = fmt.Fprintf(b, format, args...)
	b.WriteByte('\n')
}

func formatDay(day time.Time) string {
	if day.IsZero() {
		return "n/a"
	}
	return day.Format("2006-01-02")
}

func formatHour(hour int) string {
	if hour < 0 || hour > 23 {
		return "n/a"
	}
	return fmt.Sprintf("%02d:00", hour)
}

func formatWeekdays(counts [7]int) string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	parts := make([]string, 0, len(names))
	for i, name := range names {
		parts = append(parts, fmt.Sprintf("%s=%d", name, counts[i]))
	}
	return strings.Join(parts, ", ")
}

func formatDurationOrNA(value time.Duration, count int) string {
	if count == 0 {
		return "n/a"
	}
	return formatDuration(value)
}

func formatDuration(value time.Duration) string {
	value = value.Round(time.Second)
	hours := int(value / time.Hour)
	value -= time.Duration(hours) * time.Hour
	minutes := int(value / time.Minute)
	value -= time.Duration(minutes) * time.Minute
	seconds := int(value / time.Second)

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func formatParticipantTotal(total participantTotal) string {
	if total.Name == "" {
		return "n/a"
	}
	return fmt.Sprintf("%s (%d)", total.Name, total.Count)
}
