# ADR-012: One subject for each spec

- Status: Accepted
- Date: 2026-10-27
- Owners: Equal

## Context
The spec locks-and-cache had four subjects: plugin extraction, the cache lock and sessions, the docs lock, and the project data directory. The spec cli had all commands in one spec. Some mechanism requirements named commands (for example "the update and doctor commands hold the cache lock"). A reader could not find all the behavior of one command in one spec, and a mechanism requirement changed each time a new command used the mechanism.

## Decision
1. Each spec has one subject: one command or one mechanism.
2. A mechanism spec does not name a command. It gives how the mechanism works.
3. Only command specs give the user-facing behavior. A command spec says which mechanisms the command uses.
4. Each claude-dx command has its own spec cli-<command> (cli-launch, cli-update, cli-doctor, cli-clean). The behavior that all commands share is in cli-common.
5. The MCP server is one command spec, mcp-server. The docs-* specs give its tool areas.
6. Mechanism specs: plugin (extraction), session (cache lock and sessions), docs-common (docs lock and atomic write), project-data (project data directory, project key and number allocation).

## Alternatives
- Keep locks-and-cache and cli as they are: Rejected. One spec mixes subjects, and the behavior of one command is in many places.
- Split locks-and-cache only, and keep one cli spec: Rejected. The command parts of the mechanism requirements go to one large cli spec, and a reader still reads all commands to find one.
- Move each requirement with its command names into the new mechanism specs: Rejected. A mechanism spec then changes each time a new command uses the mechanism.

## Consequences
- Positive: A reader finds all the behavior of one command in one spec. A mechanism spec stays correct when a new command uses the mechanism. The cli-* names follow the docs-* names.
- Negative: There are more specs. A requirement that names a command and a mechanism is split into two requirements in two specs.
- Follow-up: Split cli into cli-common, cli-launch, cli-update, cli-doctor and cli-clean. Split locks-and-cache into plugin, session, docs-common and project-data. Move the update check requirement to mcp-server. Delete cli and locks-and-cache. Task general/split-locks-and-cache.

## Evidence
docs/specs/locks-and-cache/SPEC.md section 5: R-10 names the MCP server and the update and doctor commands; R-8 is the global cache removal of the clean command; R-15 names the MCP server. docs/specs/cli/SPEC.md has 21 requirements for launch, update, doctor and clean.
