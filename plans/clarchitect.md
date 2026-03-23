# Plan: clarchitect

> Source PRD: `clarchitect-prd.md` (including Implementation Decisions appendix)

## Architectural decisions

Durable decisions that apply across all phases:

- **Module path**: `github.com/fatstruct/clarchitect`, Go 1.25
- **Entry point**: `main.go` → `internal/cli/cli.go` (one-liner main, all dispatch in internal package)
- **Layers**: CLI → Prompt + Engine → Registry. Unidirectional dependencies.
- **Registry API**: Builder pattern with chainable fluent syntax (`NewStack().Label().Variant().Variable().File().Done()`)
- **Templates**: `embed.FS` via `templates/embed.go`, `text/template` rendering, flat variable model
- **Template paths**: `templates/<stack>/<variant>/` with `CLAUDE.md.tmpl` + `rules/*.md.tmpl`
- **TUI**: Bubbletea v2 custom models, lipgloss v1 + termenv for styling
- **Non-interactive mode**: TTY detection, auto-derived kebab-case flags (with registry override), `--force` for overwrite
- **File writing**: `os.MkdirAll` 0755 for dirs, 0644 for files
- **Version**: `ldflags` → `ReadBuildInfo()` → `"dev"` fallback
- **Testing**: TDD throughout — table-driven tests, `teatest` for TUI flows, `TestAllTemplatesParse` for template validation
- **Commands**: `global`, `project [variant]`, `list`, `help`, `version` (also `--version`, `-v`)

---

## Phase 1: Registry + Engine + Global Command (non-interactive)

**User stories**: Global command, template rendering, file writing with overwrite protection, non-interactive mode

### What to build

The foundational data pipeline end-to-end. Define the registry builder API with struct types for Stack, Variant, Variable, and FileMapping. Register the global "stack" with its two variables (`AuthorName`, `PreferredTestStyle`). Build the engine: parse an embedded template, render it with collected variables, write the output file to disk with `os.MkdirAll`. Wire `main.go` → `internal/cli/` with the `global` command accepting `--author-name` and `--preferred-test-style` flags. Implement `--force` for overwrite behavior. Detect non-TTY and operate in flag-only mode. Create a placeholder `templates/global/CLAUDE.md.tmpl` with template variable markers (real content comes in phase 6). Parse all templates on startup and fail fast if malformed.

### Acceptance criteria

- [ ] `go build .` produces a `clarchitect` binary
- [ ] `clarchitect global --author-name "Test"` writes `~/.claude/CLAUDE.md` with rendered template
- [ ] Engine creates `~/.claude/` directory if it doesn't exist
- [ ] Running again without `--force` skips the file with a warning
- [ ] Running again with `--force` overwrites the file
- [ ] Missing required `--author-name` in non-interactive mode exits with error
- [ ] Default value for `--preferred-test-style` is applied when flag is omitted
- [ ] Malformed template causes immediate startup failure with clear error
- [ ] Registry builder API compiles and registers the global variant
- [ ] `TestAllTemplatesParse` passes
- [ ] Unit tests for engine rendering, registry builder, flag derivation

---

## Phase 2: Project Command (non-interactive)

**User stories**: Project command with variant argument, multi-file rendering, registry extension point, flag derivation

### What to build

Extend the CLI with the `project` command. Register all four variants (`typescript-nextjs`, `typescript-express`, `swift-swiftui`, `go-chi`) in the registry with their variables and file mappings. The engine renders multiple files per variant, creating `.claude/rules/` subdirectories as needed. Each variant's variables auto-derive kebab-case flag names with optional registry override. Placeholder templates for all variants (real content in phase 6). Overwrite behavior: default skip with warnings per file, `--force` overwrites all.

### Acceptance criteria

- [ ] `clarchitect project go-chi --project-name my-api --go-module github.com/me/my-api` writes `CLAUDE.md` + `.claude/rules/*.md` in cwd
- [ ] All four variants are registered and selectable by key
- [ ] Invalid variant key exits with error listing available variants
- [ ] Multi-file rendering creates all files for a variant in one run
- [ ] `.claude/rules/` directory is created automatically
- [ ] Per-file overwrite skip/warn works without `--force`
- [ ] `--force` overwrites all existing files
- [ ] Required variables without defaults cause error when omitted
- [ ] Variables with defaults apply default when flag is omitted
- [ ] Auto-derived flag names match expected kebab-case (`GoModule` → `--go-module`, `MinIOSVersion` → `--min-ios-version`)
- [ ] Registry override for flag names works
- [ ] `TestAllTemplatesParse` covers all new templates
- [ ] Unit tests for variant lookup, multi-file engine, flag name derivation

---

## Phase 3: Interactive TUI — Global Command

**User stories**: Interactive global flow, TTY detection

### What to build

Custom bubbletea models for the interactive global flow. Text input model with placeholder text, default value acceptance on Enter, and required-field re-prompting. Overwrite confirmation model (y/n/a). TTY detection at CLI entry: if interactive, launch bubbletea; if not, fall through to flag-based mode from phase 1. Lipgloss styling for prompts, success/skip output, and error messages. Adaptive colors via termenv.

### Acceptance criteria

- [ ] Running `clarchitect global` in a terminal launches interactive prompts
- [ ] Text input shows prompt label with default value hint
- [ ] Pressing Enter on empty required field re-prompts
- [ ] Pressing Enter on field with default accepts the default
- [ ] If `~/.claude/CLAUDE.md` exists, overwrite confirmation appears (y/n/a)
- [ ] `y` overwrites current file, `n` skips with message, `a` overwrites all remaining
- [ ] Styled output: success checkmarks, skip indicators, error messages
- [ ] Colors adapt to terminal capabilities and respect `NO_COLOR`
- [ ] Non-TTY stdin still works via flags (phase 1 behavior preserved)
- [ ] `teatest` tests for the text input and overwrite confirmation flows
- [ ] Unit tests for TTY detection logic

---

## Phase 4: Interactive TUI — Project Command

**User stories**: Interactive project flow, stack selection, variant selection

### What to build

Custom bubbletea models for stack and variant selection. Numbered list model with keyboard navigation (arrow keys, j/k, number keys) and Enter to confirm. When `clarchitect project` is run without an argument: show stack selection → if stack has multiple variants, show variant selection → collect variables via text input models from phase 3 → render files with overwrite confirmation. When run with an argument: skip selection, go straight to variable collection.

### Acceptance criteria

- [ ] `clarchitect project` (no argument) shows stack selection menu
- [ ] Selecting a stack with multiple variants shows variant selection
- [ ] Selecting a stack with one variant skips to variable collection
- [ ] Keyboard navigation works: arrows, j/k, number keys, Enter
- [ ] After selection, variable prompts appear for chosen variant
- [ ] Overwrite confirmation works for each existing file
- [ ] `clarchitect project go-chi` skips selection, prompts for variables only
- [ ] Full flow renders all files with styled output
- [ ] `teatest` tests for selection menu navigation and full interactive flow

---

## Phase 5: List, Help, Version Commands

**User stories**: `list`, `help`, `version` commands

### What to build

Three non-interactive commands with lipgloss-styled output. `list` reads all stacks/variants from the registry and formats them as a styled tree with CLI usage hints. `help` prints usage synopsis, command descriptions, examples, and version. `version` prints the version string sourced from ldflags → `ReadBuildInfo()` → `"dev"`. Also wire `--version` and `-v` flags to version output.

### Acceptance criteria

- [ ] `clarchitect list` prints all stacks and variants with their `clarchitect project <key>` usage
- [ ] `clarchitect help` prints usage, all commands, examples, and version
- [ ] `clarchitect version` prints version string
- [ ] `--version` and `-v` flags print version string
- [ ] Output is lipgloss-styled with consistent visual treatment
- [ ] Output is pipeable (no bubbletea, no raw terminal mode)
- [ ] Unknown command shows help text with error
- [ ] `ldflags`-injected version takes priority over `ReadBuildInfo()`
- [ ] Unit tests for version resolution fallback chain
- [ ] Unit tests for list/help output content

---

## Phase 6: Template Content

**User stories**: All template specifications from the PRD

### What to build

Author the actual Markdown template content for all variants, replacing placeholders from earlier phases. Global template: coding principles, error handling, testing, git, code review. TypeScript+Next.js: RSC architecture, Zod, Vitest, Tailwind. TypeScript+Express: layered architecture, asyncHandler, supertest. Swift+SwiftUI: MVVM, @Observable, Swift Testing. Go+Chi: handler→service→repository, table-driven tests, httptest. Each template uses `{{.VariableName}}` for substitution. All templates must pass `TestAllTemplatesParse` and render correctly with sample variable values.

### Acceptance criteria

- [ ] Global template covers: principles, error handling, testing (uses `{{.PreferredTestStyle}}`), git, code review
- [ ] TypeScript+Next.js templates cover: architecture, commands, code style, testing per PRD spec
- [ ] TypeScript+Express templates cover: layered architecture, middleware, code style, testing per PRD spec
- [ ] Swift+SwiftUI templates cover: MVVM, SwiftUI rules, testing per PRD spec
- [ ] Go+Chi templates cover: architecture, API conventions, testing per PRD spec
- [ ] All templates render with sample variables without errors
- [ ] Rendered output is valid, well-structured Markdown
- [ ] `TestAllTemplatesParse` passes
- [ ] Integration tests: render each variant with sample data and verify output contains expected sections
