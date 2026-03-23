<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="MIT License" />
  <img src="https://img.shields.io/badge/Status-Beta-orange?style=flat-square" alt="Beta" />
</p>

<h1 align="center">clarchitect</h1>

<p align="center">
  <strong>Scaffold Claude Code configuration files from curated, embedded templates.</strong><br />
  One command for your global coding identity. One command for per-project rules.
</p>

---

## The Problem

Claude Code reads `CLAUDE.md` at the start of every session. Without one, you repeat yourself — architecture decisions, coding style, testing conventions — session after session. The workarounds are painful:

- **`/init`** generates a generic starter that needs heavy editing every time
- **Copy-paste from old projects** drifts over time
- **Nothing** — you explain the same rules in natural language each session

For developers who work across multiple stacks (TypeScript, Swift, Go), the problem compounds. Each stack has different patterns, tools, and conventions.

## The Solution

`clarchitect` is a single-binary CLI that:

1. **Bootstraps `~/.claude/CLAUDE.md`** with your universal coding identity (run once)
2. **Generates project-level config** (`CLAUDE.md` + `.claude/rules/*.md`) from stack-specific templates

Templates are embedded in the binary. Every project you scaffold gets your latest conventions automatically.

---

## Quick Start

### Install

```bash
# From source
go install github.com/fatstruct/clarchitect@latest

# Or clone and build
git clone https://github.com/fatstruct/clarchitect.git
cd clarchitect
make build
```

### Set Up Your Global Config

```bash
clarchitect global
```

This launches an interactive TUI that asks for your name and testing philosophy, then writes `~/.claude/CLAUDE.md`.

### Scaffold a Project

```bash
# Interactive — choose stack and variant from a menu
clarchitect project

# Direct — skip selection, go straight to variables
clarchitect project go-chi
```

This generates a `CLAUDE.md` and `.claude/rules/*.md` files in the current directory, customized with your project's name, module path, and framework choices.

---

## Commands

| Command | Description |
|---------|-------------|
| `clarchitect global` | Set up `~/.claude/CLAUDE.md` with your coding identity |
| `clarchitect project` | Interactive stack/variant selection + variable prompts |
| `clarchitect project <variant>` | Generate rules for a specific variant |
| `clarchitect list` | Show all available stacks and variants |
| `clarchitect version` | Print version |
| `clarchitect help` | Show help |

### Interactive Mode

When run in a terminal without flags, both `global` and `project` launch a polished TUI with keyboard navigation:

```
Select a stack
> 1. TypeScript
  2. Swift
  3. Go

  Use arrows/j/k to navigate, 1-9 or Enter to select
```

```
Select a variant
> 1. TypeScript + Next.js
  2. TypeScript + Express API

  Use arrows/j/k to navigate, 1-9 or Enter to select
```

```
Project name: my-api
Node.js version [22]: ▊
```

If a file already exists, you're prompted to overwrite:

```
File already exists: ./CLAUDE.md
Overwrite? [y]es / [n]o / [a]ll
```

### Non-Interactive Mode

Pass flags to skip the TUI entirely — perfect for CI/CD or scripting:

```bash
clarchitect global --author-name "Jane Doe" --preferred-test-style "TDD"

clarchitect project go-chi \
  --project-name my-api \
  --go-module github.com/me/my-api

clarchitect project typescript-nextjs \
  --project-name my-app \
  --package-manager pnpm \
  --force
```

Use `--force` to overwrite existing files without prompting.

---

## Available Templates

### Global

Sets up `~/.claude/CLAUDE.md` with language-agnostic engineering principles.

| Variable | Flag | Required | Default |
|----------|------|----------|---------|
| Author name | `--author-name` | Yes | — |
| Testing philosophy | `--preferred-test-style` | No | `test alongside implementation` |

**Generates:** General principles, error handling, testing philosophy, git conventions, code review mindset.

---

### TypeScript + Next.js

```bash
clarchitect project typescript-nextjs
```

| Variable | Flag | Required | Default |
|----------|------|----------|---------|
| Project name | `--project-name` | Yes | — |
| Node.js version | `--node-version` | No | `22` |
| Package manager | `--package-manager` | No | `pnpm` |
| Next.js version | `--next-version` | No | `15` |

**Generates:**
| File | Content |
|------|---------|
| `CLAUDE.md` | `src/` architecture, RSC-first patterns, commands |
| `.claude/rules/code-style.md` | Strict mode, named exports, Tailwind-only, file naming |
| `.claude/rules/testing.md` | Vitest + Testing Library, behavior-driven tests |

---

### TypeScript + Express API

```bash
clarchitect project typescript-express
```

| Variable | Flag | Required | Default |
|----------|------|----------|---------|
| Project name | `--project-name` | Yes | — |
| Node.js version | `--node-version` | No | `22` |
| Package manager | `--package-manager` | No | `pnpm` |

**Generates:**
| File | Content |
|------|---------|
| `CLAUDE.md` | Layered architecture (routes/services/repositories), middleware chain |
| `.claude/rules/code-style.md` | Strict mode, kebab-case files, Zod validation |
| `.claude/rules/testing.md` | Vitest + supertest integration tests |

---

### Swift + SwiftUI (iOS)

```bash
clarchitect project swift-swiftui
```

| Variable | Flag | Required | Default |
|----------|------|----------|---------|
| Project name | `--project-name` | Yes | — |
| Bundle identifier | `--bundle-id` | Yes | — |
| Min iOS version | `--min-ios-version` | No | `17` |

**Generates:**
| File | Content |
|------|---------|
| `CLAUDE.md` | MVVM architecture, async/await, NavigationPath routing |
| `.claude/rules/swiftui.md` | `@Observable`, `.task` lifecycle, accessibility, layout |
| `.claude/rules/testing.md` | Swift Testing framework, protocol-based mocking |

---

### Go + Chi Router

```bash
clarchitect project go-chi
```

| Variable | Flag | Required | Default |
|----------|------|----------|---------|
| Project name | `--project-name` | Yes | — |
| Go module path | `--go-module` | Yes | — |

**Generates:**
| File | Content |
|------|---------|
| `CLAUDE.md` | `cmd/`/`internal/` architecture, handler/service/repository layers |
| `.claude/rules/api-conventions.md` | RESTful URLs, handler struct pattern, cursor pagination |
| `.claude/rules/testing.md` | Table-driven tests, `httptest`, interface mocking |

---

## Example Output

Running `clarchitect project go-chi` with project name `my-api` and module `github.com/me/my-api` produces:

**`CLAUDE.md`:**
```markdown
# my-api

Go + Chi router. Module: `github.com/me/my-api`.

## Architecture

cmd/
└── server/         # Application entry point
internal/
├── handler/        # HTTP handlers
├── service/        # Business logic
├── repository/     # Data access
├── middleware/      # HTTP middleware
├── model/          # Domain types
└── config/         # Configuration loading

## Commands

go build ./cmd/server
go test ./...
go vet ./...
...
```

**`.claude/rules/api-conventions.md`** — RESTful URL patterns, handler struct with dependency injection, response helpers, middleware chain, cursor pagination.

**`.claude/rules/testing.md`** — Table-driven tests, `httptest` for handlers, mock at interfaces, test naming conventions.

---

## Development

```bash
make build       # Build binary (version from git)
make test        # Run all tests
make vet         # Static analysis
make clean       # Remove binary
make install     # Install to $GOPATH/bin

# Build with specific version
make build VERSION=1.0.0
```

### Architecture

```
main.go                          # Entry point
internal/
├── cli/cli.go                   # Command dispatch, TTY detection, flag parsing
├── engine/engine.go             # Template rendering + file writing
├── registry/registry.go         # Stack/variant/variable definitions (builder pattern)
├── tui/                         # Bubbletea v2 interactive models
│   ├── textinput.go             # Text input with validation
│   ├── confirm.go               # Overwrite confirmation (y/n/a)
│   ├── selection.go             # Numbered list with keyboard navigation
│   ├── globalform.go            # Global command orchestrator
│   ├── projectform.go           # Project command orchestrator
│   └── styles.go                # Lipgloss v2 styles
└── version/version.go           # Version resolution (ldflags → BuildInfo → "dev")
templates/                       # Embedded via go:embed
├── global/CLAUDE.md.tmpl
├── go/chi/...
├── typescript/nextjs/...
├── typescript/express/...
└── swift/swiftui/...
```

### Adding a New Template

1. Add the variant to the registry in `internal/registry/registry.go` using the builder pattern
2. Create template files under `templates/<stack>/<variant>/`
3. Run `go test ./...` — `TestAllTemplatesParse` will verify your templates compile

---

## Roadmap

- [ ] `go-gin` variant (Gin framework)
- [ ] `go-stdlib` variant (stdlib `net/http` only)
- [ ] `swift-uikit` variant (UIKit for legacy iOS)
- [ ] `typescript-hono` variant (Hono framework)
- [ ] `python-fastapi` variant
- [ ] `python-django` variant
- [ ] `rust-axum` variant
- [ ] Custom template directories (user-defined templates outside the binary)
- [ ] `clarchitect update` — re-run with new template versions, preserving user edits
- [ ] Template validation linting (`clarchitect validate`)
- [ ] Homebrew formula

---

## License

[MIT](LICENSE) - Astrocatto
