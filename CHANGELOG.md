# Changelog

## v3.0.0

- added an unlimited message-type ranking to the `--stats` TXT report
- added per-participant reaction tops to the `--stats` TXT report
- prepared release documentation for the richer chat analytics release line

## v2.0.0

- added `--activity-png FILE` to generate a PNG activity chart with daily message counts from the first exported message to the last
- kept TXT conversion and PNG chart generation in a single run from the same Telegram export
- included quiet days in the activity timeline so long gaps remain visible on the chart
- added `--stats FILE` to generate a TXT report with chat, participant, response-time, and content metrics
- documented the new flag and release scope in the README
