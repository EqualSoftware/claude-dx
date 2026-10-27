# ADR-{{printf "%03d" .Number}}: {{.Title}}

- Status: {{.Status}}
- Date: {{.Date}}
- Owners: {{.Owners}}

## Context
{{.Context}}

## Decision
{{.Decision}}

## Alternatives
{{range .Alternatives}}- {{.Option}}: {{.Reason}}
{{else}}- None.
{{end}}
## Consequences
- Positive: {{.Consequences.Positive}}
- Negative: {{.Consequences.Negative}}
- Follow-up: {{.Consequences.FollowUp}}

## Evidence
{{.Evidence}}
