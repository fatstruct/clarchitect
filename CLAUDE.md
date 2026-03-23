# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`clarchitect` — Go CLI that scaffolds Claude Code config files from embedded templates. See `clarchitectt-prd.md` for the full spec.

## Commands

```bash
go build -o clarchitect .     # build
go test ./...                # all tests
go test ./internal/engine/   # single package
go vet ./...                 # lint
```

## Architecture

CLI (`cmd/cli.go`) → Prompt (`internal/prompt/`) + Engine (`internal/engine/`) → Registry (`internal/registry/`).

- **Registry** defines stacks, variants, variables, file mappings — the only file to touch when adding variants
- **Engine** renders `text/template` templates and writes files with overwrite protection
- **Prompt** handles interactive input: variable collection, selection menus, overwrite confirm (y/n/a)
- **Templates** (`templates/<stack>/<variant>/`) are embedded via `go:embed` — no runtime FS dependency

## Constraints

- No CLI framework (manual dispatch for 5 commands), no external template engine
- Flat template variables only, no template inheritance
- Stdlib `text/template` + `embed.FS` — zero external deps for core functionality
- Table-driven tests with `testing` stdlib
