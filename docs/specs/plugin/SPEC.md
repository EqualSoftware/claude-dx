---
slug: plugin
title: 'dx plugin: skills and rules'
status: Approved
milestone: v0.1.0
priority: P1
area: [plugin, skills]
related: [ADR-002, ADR-010, ADR-012]
---

# dx plugin: skills and rules

## 1. Summary

The dx plugin is the set of files that claude-dx gives to Claude Code. It holds the skills, the shared rules text, and (in later specs) the hooks. The launcher extracts the plugin to the cache and loads it with --plugin-dir. This spec also sets how the extraction works.

## 2. Problem

A user of claude-dx needs the dx skills and rules in each session, with no setup. The skills must use the dx tools, and the rules must hold for each session.

## 3. Goals

- G-1: /dx:task, /dx:adr, /dx:spec, /dx:glossary, /dx:learning and /dx:grill are available in each session.
- G-2: The rules text is in the system prompt of each session, and it sets the rules for documents and questions.
- G-3: During work, the model asks the user before each choice that a task or ADR does not decide.

## 4. Out of Scope

- The hooks and the Bash git block: see the spec git-tools.
- The skills for git (/dx:commit) and the eureka skill: see the spec git-tools.
- The resume command and the SessionStart hook: see the spec resume-command.
- The cache lock and the removal of unused plugin directories: see the spec session.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The plugin ships the skills adr, glossary, grill, learning, spec and task. | Done |
| R-2 | The rules text rules.md is extracted to the plugin directory, and the launcher gives it with --append-system-prompt-file. It has at most 15 lines. | Done |
| R-3 | The rules say to write dx documents and dx tool arguments in ASD-STE100, to talk to the user in the user language, and to change dx documents only with the dx tools. | Done |
| R-4 | The rules say to stop and ask the user about each choice that the task or ADR does not decide, and about each conflict with an Accepted ADR or another task. | Done |
| R-5 | A list request to /dx:task runs task_list with the filters of the request, gives its lines, and runs no other tool. | Done |
| R-6 | The /dx:grill checklist covers each error path and fallback, each user-facing text, each existing command or tool that a change affects, and each conflict with an Accepted ADR. | Done |
| R-7 | The /dx:grill skill writes each answer with spec_question_answer, after it adds the question with spec_question_add. | Done |
| R-8 | The plugin has no agent (ADR-010). | Done |
| R-9 | The plugin of a release build is extracted to <cache>/plugin/<version>-<key>, where the key is 8 bytes of the SHA-256 of the binary path. Each binary path has its own directory. | Done |
| R-10 | The plugin of a dev build is extracted to <cache>/plugin/dev-<key>, where the key comes from the embedded files and the binary path. A new dev build does not change the files of a running session. | Done |
| R-11 | The plugin extraction writes to a temporary directory and renames it. Two parallel starts never leave a partial plugin. | Done |
| R-12 | The extraction replaces {{BINARY}} and {{VERSION}} in the .json files. The binary path is JSON-escaped, so Windows paths are valid. | Done |

## 6. Design

### Plugin layout (`internal/plugin/assets/`)

```text
.claude-plugin/plugin.json     name "dx" and the version
.mcp.json                      server mcp (see mcp-server)
rules.md                       the shared rules, given with --append-system-prompt-file
skills/<name>/SKILL.md         adr, glossary, grill, learning, spec, task
```

### Extraction layout

```text
<cache>/plugin/<version>-<key>/   release plugin, key of the binary path
<cache>/plugin/dev-<key>/         dev plugin, key of the files and the binary path
<cache>/plugin/*.tmp-*/           temporary extraction
```

### Modules

- `internal/plugin/plugin.go`: `DirName` and `Ensure`. `Ensure` extracts the assets to a temporary directory, replaces `{{BINARY}}` and `{{VERSION}}` in the `.json` files, and renames the directory.
- `internal/launcher/launcher.go`: `Args` gives `--plugin-dir`, `--settings` and `--append-system-prompt-file`.

### Skill rules

- Each skill uses the dx tools of its document type. A skill does not change a dx document with Edit or Write.
- The grill skill writes its answers with the spec question tools.
- A choice that a task does not decide is a question to the user, before the work (rules, task skill step 5).

### Decisions

- ADR-002: the MCP tool conventions.
- ADR-010: skills and shared rules with no agent.
- ADR-012: one subject for each spec.

## 7. Acceptance Criteria

Done requirements:
- R-1: `TestEnsureExtracts` (`internal/plugin/plugin_test.go`) checks the extraction of the plugin files.
- R-2: `TestRulesExtracted` (`internal/plugin/plugin_test.go`) and `TestArgs` (`internal/launcher/launcher_test.go`) for the file flag.
- R-3 to R-7: no automated test. Each is checked by reading the file in `internal/plugin/assets/` (`rules.md`, `skills/task/SKILL.md`, `skills/grill/SKILL.md`). For R-5, a `/dx:task` list request in a session must show `task_list` as the only tool call.
- R-8: no automated test. Check that `internal/plugin/assets/` has no `agents` folder.
- R-9: `TestEnsureReleasePerBinary`, `TestEnsureReusesRelease` (`internal/plugin/plugin_test.go`).
- R-10: `TestEnsureDevKeepsRunningSession` (`internal/plugin/plugin_test.go`).
- R-11: `TestEnsureParallel` (`internal/plugin/plugin_test.go`).
- R-12: `TestEnsureExtracts` and `TestJSONEscapeWindowsPath` (`internal/plugin/plugin_test.go`).

Planned requirements: none.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | R-8: text changed. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 5. |
| 2026-10-27 | Equal | Added R-10. |
| 2026-10-27 | Equal | Added R-11. |
| 2026-10-27 | Equal | Added R-12. |
| 2026-10-27 | Equal | R-9: status Done. |
| 2026-10-27 | Equal | R-10: status Done. |
| 2026-10-27 | Equal | R-11: status Done. |
| 2026-10-27 | Equal | R-12: status Done. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | Approved by the user after the split of locks-and-cache (ADR-012). |