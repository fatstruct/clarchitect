package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
	"github.com/fatstruct/clarchitect/internal/engine"
	"github.com/fatstruct/clarchitect/internal/registry"
)

// testStacks returns a set of stacks for testing purposes.
func testStacks() []registry.Stack {
	return []registry.Stack{
		{
			Key:   "go",
			Label: "Go",
			Variants: []registry.Variant{
				{
					Key:        "go-chi",
					Label:      "Go + Chi Router",
					StackKey:   "go",
					VariantDir: "chi",
					Vars: []registry.Variable{
						{Key: "ProjectName", Prompt: "Project name", Default: ""},
						{Key: "GoModule", Prompt: "Go module path", Default: ""},
					},
					Files: []registry.FileMapping{
						{TemplatePath: "CLAUDE.md.tmpl", OutputPath: "CLAUDE.md"},
					},
				},
			},
		},
		{
			Key:   "typescript",
			Label: "TypeScript",
			Variants: []registry.Variant{
				{
					Key:        "typescript-nextjs",
					Label:      "TypeScript + Next.js",
					StackKey:   "typescript",
					VariantDir: "nextjs",
					Vars: []registry.Variable{
						{Key: "ProjectName", Prompt: "Project name", Default: ""},
					},
					Files: []registry.FileMapping{
						{TemplatePath: "CLAUDE.md.tmpl", OutputPath: "CLAUDE.md"},
					},
				},
				{
					Key:        "typescript-express",
					Label:      "TypeScript + Express API",
					StackKey:   "typescript",
					VariantDir: "express",
					Vars: []registry.Variable{
						{Key: "ProjectName", Prompt: "Project name", Default: ""},
					},
					Files: []registry.FileMapping{
						{TemplatePath: "CLAUDE.md.tmpl", OutputPath: "CLAUDE.md"},
					},
				},
			},
		},
	}
}

// testEngine creates a test engine with a simple template.
func testEngine() *engine.Engine {
	fsys := fstest.MapFS{
		"go/chi/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.ProjectName}}\nModule: {{.GoModule}}\n"),
		},
		"typescript/nextjs/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.ProjectName}}\n"),
		},
		"typescript/express/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.ProjectName}}\n"),
		},
	}
	eng := engine.New(fsys, ".")
	if err := eng.ParseTemplates(); err != nil {
		panic("failed to parse test templates: " + err.Error())
	}
	return eng
}

func TestProjectFormStackSelection(t *testing.T) {
	stacks := testStacks()
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectForm(stacks, eng, dir)
	var model tea.Model = m

	// Should show stack selection initially
	pm := model.(ProjectFormModel)
	if pm.Phase() != projectPhaseSelectStack {
		t.Fatalf("Phase() = %d, want projectPhaseSelectStack", pm.Phase())
	}

	view := pm.viewString()
	if !strings.Contains(view, "Select a stack") {
		t.Errorf("expected stack selection title in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Go") {
		t.Errorf("expected 'Go' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "TypeScript") {
		t.Errorf("expected 'TypeScript' in view, got:\n%s", view)
	}
}

func TestProjectFormSingleVariantStack(t *testing.T) {
	stacks := testStacks()
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectForm(stacks, eng, dir)
	var model tea.Model = m

	// Select "Go" (index 0) — has only one variant, should skip variant selection
	var cmd tea.Cmd
	model, cmd = sendRune(model, '1')

	// Process the SelectionDoneMsg
	if cmd != nil {
		model, _ = model.Update(cmd())
	}

	pm := model.(ProjectFormModel)
	if pm.Phase() != projectPhaseCollectVars {
		t.Fatalf("Phase() = %d, want projectPhaseCollectVars (should skip variant selection)", pm.Phase())
	}
	if pm.Variant().Key != "go-chi" {
		t.Errorf("Variant().Key = %q, want %q", pm.Variant().Key, "go-chi")
	}
}

func TestProjectFormMultiVariantStack(t *testing.T) {
	stacks := testStacks()
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectForm(stacks, eng, dir)
	var model tea.Model = m

	// Select "TypeScript" (index 1) — has two variants, should show variant selection
	var cmd tea.Cmd
	model, cmd = sendRune(model, '2')

	// Process the SelectionDoneMsg
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	pm := model.(ProjectFormModel)
	if pm.Phase() != projectPhaseSelectVariant {
		t.Fatalf("Phase() = %d, want projectPhaseSelectVariant", pm.Phase())
	}

	view := pm.viewString()
	if !strings.Contains(view, "Select a variant") {
		t.Errorf("expected variant selection title in view, got:\n%s", view)
	}
	if !strings.Contains(view, "TypeScript + Next.js") {
		t.Errorf("expected Next.js variant in view, got:\n%s", view)
	}
	if !strings.Contains(view, "TypeScript + Express API") {
		t.Errorf("expected Express variant in view, got:\n%s", view)
	}

	// Select "TypeScript + Express API" (index 1)
	model, cmd = sendRune(model, '2')
	if cmd != nil {
		model, _ = model.Update(cmd())
	}

	pm = model.(ProjectFormModel)
	if pm.Phase() != projectPhaseCollectVars {
		t.Fatalf("Phase() = %d, want projectPhaseCollectVars after variant selection", pm.Phase())
	}
	if pm.Variant().Key != "typescript-express" {
		t.Errorf("Variant().Key = %q, want %q", pm.Variant().Key, "typescript-express")
	}
}

func TestProjectFormWithVariantSkipsSelection(t *testing.T) {
	variant := testStacks()[0].Variants[0] // go-chi
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectFormWithVariant(variant, eng, dir)

	if m.Phase() != projectPhaseCollectVars {
		t.Fatalf("Phase() = %d, want projectPhaseCollectVars", m.Phase())
	}

	view := m.viewString()
	if !strings.Contains(view, "Project name") {
		t.Errorf("expected 'Project name' prompt in view, got:\n%s", view)
	}
}

func TestProjectFormWithVariantFullFlow(t *testing.T) {
	variant := testStacks()[0].Variants[0] // go-chi
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectFormWithVariant(variant, eng, dir)
	var model tea.Model = m

	// Type project name
	model = typeString(model, "my-api")
	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)

	// Process TextInputDoneMsg
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	// Now should prompt for Go module path
	pm := model.(ProjectFormModel)
	view := pm.viewString()
	if !strings.Contains(view, "Go module path") {
		t.Errorf("expected 'Go module path' prompt, got:\n%s", view)
	}

	// Type module path
	model = typeString(model, "github.com/me/my-api")
	model, cmd = sendKey(model, tea.KeyEnter)

	// Process TextInputDoneMsg — this should trigger file processing
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	pm = model.(ProjectFormModel)
	if pm.Phase() != projectPhaseDone {
		t.Fatalf("Phase() = %d, want projectPhaseDone", pm.Phase())
	}

	// Verify file was written
	outcomes := pm.Outcomes()
	if len(outcomes) != 1 {
		t.Fatalf("expected 1 outcome, got %d", len(outcomes))
	}
	if !outcomes[0].written {
		t.Error("expected file to be written")
	}

	// Verify file content
	content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(content), "my-api") {
		t.Errorf("expected file to contain 'my-api', got: %s", content)
	}
	if !strings.Contains(string(content), "github.com/me/my-api") {
		t.Errorf("expected file to contain module path, got: %s", content)
	}
}

func TestProjectFormOverwriteConfirmation(t *testing.T) {
	variant := testStacks()[0].Variants[0] // go-chi
	eng := testEngine()
	dir := t.TempDir()

	// Create existing file
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("existing"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	m := NewProjectFormWithVariant(variant, eng, dir)
	var model tea.Model = m

	// Fill in project name
	model = typeString(model, "my-api")
	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	// Fill in module path
	model = typeString(model, "github.com/me/my-api")
	model, cmd = sendKey(model, tea.KeyEnter)
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	// Should be in confirm overwrite phase
	pm := model.(ProjectFormModel)
	if pm.Phase() != projectPhaseConfirmOverwrite {
		t.Fatalf("Phase() = %d, want projectPhaseConfirmOverwrite", pm.Phase())
	}

	view := pm.viewString()
	if !strings.Contains(view, "CLAUDE.md") {
		t.Errorf("expected confirm view to show filename, got:\n%s", view)
	}

	// Confirm overwrite with 'y'
	model, cmd = sendRune(model, 'y')
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	pm = model.(ProjectFormModel)
	if pm.Phase() != projectPhaseDone {
		t.Fatalf("Phase() = %d, want projectPhaseDone after confirmation", pm.Phase())
	}

	outcomes := pm.Outcomes()
	if len(outcomes) != 1 {
		t.Fatalf("expected 1 outcome, got %d", len(outcomes))
	}
	if !outcomes[0].written {
		t.Error("expected file to be written after confirmation")
	}
}

func TestProjectFormDoneViewShowsOutcomes(t *testing.T) {
	variant := testStacks()[0].Variants[0] // go-chi
	eng := testEngine()
	dir := t.TempDir()
	m := NewProjectFormWithVariant(variant, eng, dir)
	var model tea.Model = m

	// Fill in both variables
	model = typeString(model, "my-api")
	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}
	model = typeString(model, "github.com/me/my-api")
	model, cmd = sendKey(model, tea.KeyEnter)
	if cmd != nil {
		model, cmd = model.Update(cmd())
	}

	pm := model.(ProjectFormModel)
	view := pm.viewString()
	if !strings.Contains(view, "CLAUDE.md") {
		t.Errorf("expected done view to show output file path, got:\n%s", view)
	}
}
