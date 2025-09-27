# Documentation Strategy

This document outlines the documentation plan for AWS Profile Manager — what exists, what's missing, who it's for, and what to write next.

---

## Current State

### What Exists

| File | Audience | Status |
|------|----------|--------|
| `README.md` | End users | Complete, but references missing docs |
| `DEVELOPER.md` | Contributors | Good overview, setup instructions |
| `docs/TestingGuide.md` | Contributors | Good test runner reference |
| `internal/sync/ReadMe.md` | Package users | Stub — links to missing `docs/packages/sync/` |
| `internal/task/ReadMe.md` | Package users | Stub — links to missing `docs/packages/task/ReadMe.md` |
| Inline godoc comments | Developers | Good coverage in most packages |

### Broken References (README)

The README links to these files that do not yet exist:
- `docs/SyncGuide.md` → linked from the Sync section
- `docs/ConfigurationGuide.md` → linked from the Configuration section

These are the highest-priority gaps because users will hit them immediately.

---

## Audiences

There are three distinct audiences, each needing different information:

### 1. End Users
People who install and run the CLI or GUI to manage their AWS profiles. They need:
- What the tool does and why it's useful
- How to install it
- How to use each command (with real examples)
- What the output means

### 2. Configuration Authors
AWS platform/DevOps engineers who write and distribute the JSON config file to their teams. They need:
- Full reference for every field in the config schema
- How to set up sync so their team always has the latest config
- Examples for common setups (SSO, IAM, multi-partition)

### 3. Contributors
Developers working on the codebase. They need:
- Project architecture and package responsibilities
- Build and test workflow
- Package-level API docs for complex packages (sync, task, profiles)
- Coding conventions (captured in `.github/copilot-instructions.md`)

---

## Documentation Plan

The documents are grouped by audience and ordered by priority.

### Priority 1 — Fix Broken Links (End Users & Config Authors)

These docs are already linked from the README and must exist.

#### `docs/ConfigurationGuide.md`
**Audience**: Configuration authors  
**Goal**: Complete reference for the JSON config file format  
**Contents**:
- Top-level schema (version, managed, unmanaged sections)
- Organization block (alias, name, description, partitions)
- Partition block (url, default_region, regions, accounts, roles)
- Account block (alias, name, id)
- IAM user block (access_key, secret_key, region, credential_process)
- AssumeRole chain block (role_arn, source_profile, role_session_name)
- Generic profile block (name, properties map)
- Annotated full example
- Common patterns (multi-org, GovCloud, mixed SSO + IAM)

#### `docs/SyncGuide.md`
**Audience**: Configuration authors (setting up sync for a team)  
**Goal**: Step-by-step guide to hosting and consuming a centralized config  
**Contents**:
- What sync does and when to use it
- Strategy comparison (HTTP vs S3 vs local)
- HTTP setup: hosting the config on a web server or CDN
- S3 setup: bucket policy, IAM permissions, SSO vs IAM auth
- Local setup: for development and testing
- Application settings: how users configure which source to use
- Caching behavior and offline operation
- Troubleshooting (wrong URL, auth failures, stale cache)

---

### Priority 2 — Developer Package Docs

These are referenced from in-package stubs and needed by contributors.

#### `docs/packages/sync/ReadMe.md`
**Audience**: Contributors  
**Goal**: How the sync package works internally  
**Contents**:
- Package purpose and responsibilities
- Architecture: fetcher interface, router, cache layer
- Supported strategies and their fetcher implementations
- Config struct and how to populate it
- API: `Sync()`, `ConfigFromSettings()`
- Adding a new sync strategy
- Error handling and validation

#### `docs/packages/sync/TestingStrategy.md`
**Audience**: Contributors  
**Goal**: Explain unit vs E2E test split for sync  
**Contents**:
- Why the split exists (external dependencies)
- What unit tests cover vs what E2E tests cover
- How to run each category
- When to add unit vs E2E tests

#### `docs/packages/sync/TestingE2e.md`
**Audience**: Contributors  
**Goal**: How to run E2E sync tests  
**Contents**:
- Required environment variables per strategy
- Setup for HTTP, S3, local
- Running with build tags
- Expected results and how to verify

#### `docs/packages/task/ReadMe.md`
**Audience**: Contributors  
**Goal**: How to use the task package in business logic  
**Contents**:
- What the task package solves (CLI/GUI agnostic progress)
- Task types: SubprocessTask, FunctionTask
- Reporter types: CliReporter, NoOpReporter, ChannelReporter
- Usage pattern in business logic functions
- Usage pattern in CLI commands
- Usage pattern in GUI handlers
- Context cancellation

---

### Priority 3 — Architecture Overview

#### `docs/Architecture.md`
**Audience**: Contributors  
**Goal**: High-level map of how all packages connect  
**Contents**:
- Layer diagram (CLI/GUI → package APIs → core packages)
- Package responsibility table
- Data flow for the main operations (install, sync, export, import)
- Naming conventions (profile names, session names)
- Key design decisions and why (dependency injection, task reporter pattern)

---

### Out of Scope (For Now)

The following are intentionally excluded to avoid over-documentation:

| Topic | Reason |
|-------|--------|
| Per-package godoc pages | Already covered well by inline comments; `go doc` serves this |
| `internal/generators` package doc | Pure content generation; API is self-explanatory |
| `internal/logging` package doc | Simple; inline comments are sufficient |
| `internal/test` package doc | Only used in tests; covered by TestingGuide |
| `internal/bundled` package doc | Build artifact; no user-facing behavior |
| GUI component internals | Fyne widgets; low value for current contributor base |

---

## Content Guidelines

### Length
- User docs: Keep it short. Use examples over prose. Trust the reader to experiment.
- Developer docs: Be thorough where the code is complex (sync, task, profiles). Skip obvious things.
- One concept per section. Avoid padding.

### Format
- Use code blocks for all commands, config snippets, and Go examples.
- Use tables for reference material (flags, fields, strategies).
- Use callout blocks (`> **Note:**`) sparingly — only for genuine gotchas.
- Lead with the "why" before the "how".

### File Naming
PascalCase for all documentation files (e.g., `ConfigurationGuide.md`, not `configuration-guide.md` or `CONFIGURATION.md`).

### What to Avoid
- Repeating information already visible in `--help` output
- Documenting internal implementation details in user-facing docs
- Listing every edge case in introductory docs (save for troubleshooting sections)
- Duplicating content across files — link instead

---

## Implementation Order

1. ✅ `docs/ConfigurationGuide.md` — fixes a broken README link, highest user impact
2. ✅ `docs/SyncGuide.md` — fixes the other broken README link
3. ✅ `docs/packages/task/ReadMe.md` — small, high value for contributors
4. ✅ `docs/packages/sync/ReadMe.md` — medium effort, needed for sync contributor work
5. ✅ `docs/packages/sync/TestingStrategy.md` and `TestingE2e.md` — low priority, mostly bookkeeping
6. ✅ `docs/Architecture.md` — useful but can wait until the codebase stabilises further
