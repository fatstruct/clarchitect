# PRD: clarchitec

A CLI tool that scaffolds Claude Code configuration files from curated, version-controlled templates. One command to set up your global coding identity, one command to stamp out per-project `CLAUDE.md` and `.claude/rules/` for any stack you work with.

---

## Problem

Claude Code reads a `CLAUDE.md` file at the start of every session. Without one, you repeat yourself — architecture decisions, coding style, testing conventions, commands to run — session after session. The current workarounds are:

- **`/init`**: Generates a starter file from project detection, but it's generic and needs heavy editing every time.
- **Copy-paste from old projects**: Drifts over time. You end up with slightly different conventions in every repo.
- **Nothing**: You explain the same rules in natural language at the start of each session.

For developers who work across multiple stacks (e.g., TypeScript for web, Swift for iOS, Go for backend), the problem is worse — each stack has different architectural patterns, conventions, and tooling. There's no centralized, reusable way to define and apply these.

---

## Solution

`clarchitec` is a single-binary CLI that:

1. Bootstraps `~/.claude/CLAUDE.md` with your universal coding identity (run once).
2. Generates project-level `CLAUDE.md` + `.claude/rules/*.md` from stack-specific templates with variable substitution.

Templates are embedded in the binary. You maintain them in the tool's repo. Every new project you scaffold gets your latest conventions automatically.

---

## User Personas

**Primary**: A developer who uses Claude Code daily across 2–4 tech stacks (TypeScript, Swift, Go). They want consistent, high-quality Claude behavior without manual setup on each new repo.

**Secondary**: A team lead who wants to distribute shared Claude Code conventions to their team by forking and customizing the tool's templates.

---

## Core Concepts

### The Configuration Hierarchy

Claude Code loads configuration in layers. `clarchitec` targets two of them:

| Layer           | Path                   | Scope                    | `clarchitec` command         |
| --------------- | ---------------------- | ------------------------ | ---------------------------- |
| Global personal | `~/.claude/CLAUDE.md`  | All projects, all stacks | `clarchitec global`          |
| Project root    | `./CLAUDE.md`          | This repo only           | `clarchitec project <stack>` |
| Modular rules   | `./.claude/rules/*.md` | This repo, per-concern   | `clarchitec project <stack>` |

The global file holds language-agnostic engineering principles. The project file holds stack-specific architecture, commands, and patterns. The rules files break cross-cutting concerns (testing, code style, API conventions) into focused, single-purpose documents.

### Stacks and Variants

A **stack** is a language or platform (e.g., `typescript`, `swift`, `go`). A **variant** is a specific framework or architectural pattern within that stack (e.g., `typescript-nextjs`, `typescript-express`, `go-chi`). This two-level hierarchy supports the user's need for different architectural decisions per project type while keeping the template set extensible.

Initial stacks and variants:

- **typescript**: `typescript-nextjs`, `typescript-express`
- **swift**: `swift-swiftui`
- **go**: `go-chi`

The user will add more variants over time (e.g., `go-gin`, `swift-uikit`, `typescript-hono`).

### Template Variables

Templates support Go `text/template` variables (e.g., `{{.ProjectName}}`, `{{.GoModule}}`). When a user runs `clarchitec project <stack>`, they are prompted interactively for each variable. Each variable can have a default value; if one exists, pressing Enter accepts it. Variables with no default are required and the user is re-prompted until they provide a value.

---

## Commands

### `clarchitec global`

Sets up `~/.claude/CLAUDE.md` with your universal coding preferences.

**Behavior:**

1. Prompt for global variables (name, testing philosophy).
2. If `~/.claude/CLAUDE.md` already exists, ask: overwrite / skip / overwrite all.
3. Render template and write to `~/.claude/CLAUDE.md`.

**Variables:**

| Variable             | Prompt             | Default                         |
| -------------------- | ------------------ | ------------------------------- |
| `AuthorName`         | Your name          | (required)                      |
| `PreferredTestStyle` | Testing philosophy | `test alongside implementation` |

**Output:** A single `~/.claude/CLAUDE.md` file containing language-agnostic principles: code style philosophy, error handling, testing approach, git conventions, code review mindset.

### `clarchitec project [stack-variant]`

Generates `CLAUDE.md` and `.claude/rules/` in the current working directory.

**With argument** (e.g., `clarchitec project go-chi`):

1. Look up the variant by key.
2. Prompt for that variant's variables.
3. Render all template files into the cwd.

**Without argument** (interactive mode):

1. Present a numbered list of stacks.
2. If the selected stack has multiple variants, present a numbered list of variants.
3. Prompt for variables.
4. Render files.

**Overwrite behavior:** For each output file that already exists, prompt: `[y]es / [n]o / [a]ll`. "All" applies to remaining files in this run only.

### `clarchitec list`

Prints all available stacks and variants with their CLI keys.

```
Available stacks:

  typescript
    TypeScript + Next.js          clarchitec project typescript-nextjs
    TypeScript + Express API      clarchitec project typescript-express

  swift
    Swift + SwiftUI (iOS)         clarchitec project swift-swiftui

  go
    Go + Chi Router               clarchitec project go-chi
```

### `clarchitec help`

Prints usage, examples, and version.

### `clarchitec version`

Prints version string. Also accessible as `--version` or `-v`.

---

## Template Specifications

### Global Template

**File:** `templates/global/CLAUDE.md.tmpl`

**Content sections:**

- General principles (composition over inheritance, pure functions, strict typing, DRY/KISS, function size limits, comment philosophy)
- Error handling (explicit handling, typed errors, actionable messages)
- Testing (per user's preferred style, happy path + edge case, behavior-named tests)
- Git (conventional commits, atomic commits, never commit secrets)
- Code review mindset (readability over cleverness, clarify before changing architecture)

### TypeScript + Next.js

**Files:**

- `CLAUDE.md.tmpl` — stack, architecture (`src/app/`, `src/components/`, `src/lib/`, `src/server/`, `src/types/`), commands, key decisions (RSC-first, Zod validation, absolute imports), patterns
- `rules/code-style.md.tmpl` — strict mode, named exports, props as type aliases, Tailwind-only CSS
- `rules/testing.md.tmpl` — Vitest, Testing Library, behavior-named tests, no snapshots

**Variables:** `ProjectName`, `NodeVersion` (default: 22), `PackageManager` (default: pnpm), `NextVersion` (default: 15)

### TypeScript + Express

**Files:**

- `CLAUDE.md.tmpl` — stack, layered architecture (routes → services → repositories), middleware, commands, patterns (Zod validation, typed AppError, asyncHandler)
- `rules/code-style.md.tmpl` — strict mode, named exports, kebab-case files, Zod colocated with routes
- `rules/testing.md.tmpl` — Vitest, supertest integration tests, mock at boundaries

**Variables:** `ProjectName`, `NodeVersion` (default: 22), `PackageManager` (default: pnpm)

### Swift + SwiftUI

**Files:**

- `CLAUDE.md.tmpl` — stack, MVVM architecture (Views/ViewModels/Models/Services/Navigation), xcodebuild command, key decisions (async/await, protocol-based DI, centralized navigation)
- `rules/swiftui.md.tmpl` — views as functions of state, @Observable over ObservableObject, .task over .onAppear, accessibility requirements, animation rules
- `rules/testing.md.tmpl` — Swift Testing framework, protocol-based mocking, behavior-named tests

**Variables:** `ProjectName`, `BundleID` (required), `MinIOSVersion` (default: 17)

### Go + Chi

**Files:**

- `CLAUDE.md.tmpl` — stack, architecture (cmd/internal/pkg, handler→service→repository), commands, key decisions (interfaces in/structs out, context-first, slog logging)
- `rules/api-conventions.md.tmpl` — REST naming, handler struct pattern, response helpers, middleware chain, cursor pagination, string UUIDs
- `rules/testing.md.tmpl` — table-driven tests, httptest for handlers, mock at interfaces

**Variables:** `ProjectName`, `GoModule` (required)

---

## Project Structure

```
clarchitec/
├── cmd/
│   └── cli.go                  # Command dispatcher (global, project, list, help, version)
├── internal/
│   ├── engine/
│   │   └── engine.go           # Template rendering + file writing with overwrite protection
│   ├── prompt/
│   │   └── prompt.go           # Interactive input: variable collection, selection, overwrite confirm
│   └── registry/
│       └── registry.go         # Stack/variant/variable/file definitions (the "manifest")
├── templates/
│   ├── embed.go                # go:embed directive for all template files
│   ├── global/
│   │   └── CLAUDE.md.tmpl
│   ├── typescript/
│   │   ├── nextjs/
│   │   │   ├── CLAUDE.md.tmpl
│   │   │   └── rules/
│   │   │       ├── code-style.md.tmpl
│   │   │       └── testing.md.tmpl
│   │   └── express/
│   │       ├── CLAUDE.md.tmpl
│   │       └── rules/
│   │           ├── code-style.md.tmpl
│   │           └── testing.md.tmpl
│   ├── swift/
│   │   └── swiftui/
│   │       ├── CLAUDE.md.tmpl
│   │       └── rules/
│   │           ├── swiftui.md.tmpl
│   │           └── testing.md.tmpl
│   └── go/
│       └── chi/
│           ├── CLAUDE.md.tmpl
│           └── rules/
│               ├── api-conventions.md.tmpl
│               └── testing.md.tmpl
├── go.mod
└── main.go
```

---

## Technical Decisions

| Decision         | Choice                        | Rationale                                                                      |
| ---------------- | ----------------------------- | ------------------------------------------------------------------------------ |
| Language         | Go                            | Single static binary, `embed.FS` for baking templates, author already knows Go |
| Template engine  | `text/template` (stdlib)      | No external deps. Sufficient for variable substitution.                        |
| CLI framework    | None (manual dispatch)        | Only 5 commands. No need for cobra/urfave overhead.                            |
| Template storage | `embed.FS`                    | Templates ship inside the binary. No runtime filesystem dependency.            |
| Distribution     | `go install` or manual binary | Standard Go toolchain. No package manager needed.                              |

---

## Adding a New Variant

The process for adding a new variant (e.g., `go-gin`) is:

1. Create `templates/go/gin/CLAUDE.md.tmpl` and `templates/go/gin/rules/*.md.tmpl`.
2. In `registry.go`, add a new `Variant` entry under the `go` stack with its key (`go-gin`), label, variables, and file mappings.
3. Rebuild. The new variant appears in `clarchitec list` and interactive selection.

No changes needed to the engine, prompt, or CLI layers. The registry is the only file that needs updating.

---

## Overwrite Behavior

When rendering files, for each destination path:

1. Check if the file already exists.
2. If it does, prompt: `File already exists: ./CLAUDE.md — Overwrite? [y]es / [n]o / [a]ll`
3. `y` — overwrite this file, continue prompting for next.
4. `n` — skip this file, print `⏭ Skipped CLAUDE.md`, continue.
5. `a` — overwrite this file and all remaining files without prompting again.

This applies to both `global` and `project` commands.

---

## Output Format

All rendered files are plain Markdown. No YAML frontmatter, no JSON config. The output of `clarchitec project go-chi` on a fresh directory looks like:

```
Scaffolding project config: Go + Chi Router

  Project name: my-api
  Go module path (e.g., github.com/you/project): github.com/me/my-api

  ✓  CLAUDE.md
  ✓  .claude/rules/api-conventions.md
  ✓  .claude/rules/testing.md

Done.
```

---

## Non-Goals (v0.1)

- **No config file for the tool itself.** Stacks, variants, and variables are defined in Go code. If the tool grows, this can move to a YAML/TOML manifest later.
- **No remote templates.** Everything is embedded. Fetching from GitHub or a registry is a future concern.
- **No project structure scaffolding.** `clarchitec` only generates Claude Code config files. It does not create `src/`, `cmd/`, package.json, or any application code.
- **No template inheritance or composition.** Each variant's templates are standalone. Shared patterns are handled by copy-pasting between template files. If this becomes painful, a `_shared/` partial system can be added later.
- **No `update` or `diff` command.** If templates change, the user re-runs `clarchitec project` and uses the overwrite prompt. Merging upstream template changes into customized local files is out of scope.

---

## Future Considerations

These are not committed — just plausible next steps if the tool proves useful:

- **`clarchitec add-rule <name>`**: Scaffold a new `.claude/rules/<name>.md` from a blank or category-specific template.
- **Template inheritance**: A `_shared/testing-base.md.tmpl` partial that stack-specific testing rules can include, reducing duplication across variants.
- **External template directory**: Support `--templates-dir ~/.clarchitec-templates/` for users who want to maintain templates outside the binary.
- **Dry-run mode**: `--dry-run` flag that prints what would be generated without writing anything.
- **Team distribution**: A companion `clarchitec pull <git-url>` that fetches a team's template repo and installs it as the template source.
