---
topic: go-yaml-flow-lists
description: How to write a YAML flow list with go.yaml.in/yaml/v3 so that items with commas stay one item. Use it when code writes frontmatter lists.
---

- A flow list made by hand with "[" + strings.Join(items, ", ") + "]" breaks an item that has a comma: `Windows (amd64, arm64)` becomes two items when YAML reads it again.
- Encode the slice into a `yaml.Node`, set `node.Style = yaml.FlowStyle`, then `yaml.Marshal(&node)`. The encoder quotes only the items that need it, example `[macOS, 'Windows (amd64, arm64)', 'a: b']`, and an empty slice gives `[]`. Evidence: `internal/markdown/frontmatter.go` `format`, docs-task#R-2.
