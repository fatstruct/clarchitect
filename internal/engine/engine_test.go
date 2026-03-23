package engine

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatstruct/clarchitect/internal/registry"
	"github.com/fatstruct/clarchitect/templates"
)

func newTestFS() fs.FS {
	return fstest.MapFS{
		"testdata/example/hello.md.tmpl": &fstest.MapFile{
			Data: []byte("Hello, {{.Name}}!\n"),
		},
		"testdata/example/multi.md.tmpl": &fstest.MapFile{
			Data: []byte("Project: {{.ProjectName}}\nModule: {{.GoModule}}\n"),
		},
	}
}

func newMalformedFS() fs.FS {
	return fstest.MapFS{
		"bad/broken.md.tmpl": &fstest.MapFile{
			Data: []byte("Hello, {{.Name"),
		},
	}
}

func TestParseTemplatesSuccess(t *testing.T) {
	eng := New(newTestFS(), "testdata")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	names := eng.TemplateNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 templates, got %d: %v", len(names), names)
	}
}

func TestParseTemplatesMalformed(t *testing.T) {
	eng := New(newMalformedFS(), "bad")
	err := eng.ParseTemplates()
	if err == nil {
		t.Fatal("expected error for malformed template, got nil")
	}
	if !strings.Contains(err.Error(), "parsing template") {
		t.Errorf("error message = %q, expected it to contain 'parsing template'", err.Error())
	}
}

func TestRenderWithVariables(t *testing.T) {
	eng := New(newTestFS(), "testdata")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	content, err := eng.Render("example/hello.md.tmpl", map[string]string{
		"Name": "World",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expected := "Hello, World!\n"
	if content != expected {
		t.Errorf("Render() = %q, want %q", content, expected)
	}
}

func TestRenderMultipleVariables(t *testing.T) {
	eng := New(newTestFS(), "testdata")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	content, err := eng.Render("example/multi.md.tmpl", map[string]string{
		"ProjectName": "my-api",
		"GoModule":    "github.com/me/my-api",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expected := "Project: my-api\nModule: github.com/me/my-api\n"
	if content != expected {
		t.Errorf("Render() = %q, want %q", content, expected)
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	eng := New(newTestFS(), "testdata")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	_, err := eng.Render("nonexistent.tmpl", map[string]string{})
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
}

func TestWriteFileCreatesDirectoriesAndFile(t *testing.T) {
	dir := t.TempDir()
	eng := New(newTestFS(), "testdata")

	path := filepath.Join(dir, "sub", "dir", "file.md")
	err := eng.WriteFile(path, "hello", false)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("file content = %q, want %q", string(data), "hello")
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// On macOS/Linux, check that file is 0644
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("file permissions = %o, want 0644", perm)
	}
}

func TestWriteFileSkipsExistingWhenNotForced(t *testing.T) {
	dir := t.TempDir()
	eng := New(newTestFS(), "testdata")

	path := filepath.Join(dir, "existing.md")
	// Create the file first
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	err := eng.WriteFile(path, "new content", false)
	if err != ErrFileExists {
		t.Errorf("WriteFile() error = %v, want ErrFileExists", err)
	}

	// Verify original content is preserved
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("file content = %q, want %q", string(data), "original")
	}
}

func TestWriteFileOverwritesWhenForced(t *testing.T) {
	dir := t.TempDir()
	eng := New(newTestFS(), "testdata")

	path := filepath.Join(dir, "existing.md")
	// Create the file first
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	err := eng.WriteFile(path, "new content", true)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("file content = %q, want %q", string(data), "new content")
	}
}

func TestRenderAndWriteAll(t *testing.T) {
	testFS := fstest.MapFS{
		"mystack/v1/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.ProjectName}}\n"),
		},
		"mystack/v1/rules/testing.md.tmpl": &fstest.MapFile{
			Data: []byte("## Testing for {{.ProjectName}}\n"),
		},
	}

	eng := New(testFS, ".")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	variant := registry.Variant{
		Key:        "mystack-v1",
		Label:      "My Stack V1",
		StackKey:   "mystack",
		VariantDir: "v1",
		Files: []registry.FileMapping{
			{TemplatePath: "CLAUDE.md.tmpl", OutputPath: "CLAUDE.md"},
			{TemplatePath: "rules/testing.md.tmpl", OutputPath: ".claude/rules/testing.md"},
		},
	}

	dir := t.TempDir()
	vars := map[string]string{"ProjectName": "test-project"}

	results, err := eng.RenderAndWriteAll(variant, vars, dir, false)
	if err != nil {
		t.Fatalf("RenderAndWriteAll() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Written {
			t.Errorf("expected %s to be written", r.OutputPath)
		}
	}

	// Verify file contents
	claudeContent, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	if string(claudeContent) != "# test-project\n" {
		t.Errorf("CLAUDE.md content = %q, want %q", string(claudeContent), "# test-project\n")
	}

	testingContent, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "testing.md"))
	if err != nil {
		t.Fatalf("reading testing.md: %v", err)
	}
	if string(testingContent) != "## Testing for test-project\n" {
		t.Errorf("testing.md content = %q, want %q", string(testingContent), "## Testing for test-project\n")
	}
}

func TestRenderAndWriteAllSkipsExisting(t *testing.T) {
	testFS := fstest.MapFS{
		"mystack/v1/CLAUDE.md.tmpl": &fstest.MapFile{
			Data: []byte("# {{.ProjectName}}\n"),
		},
	}

	eng := New(testFS, ".")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() error = %v", err)
	}

	variant := registry.Variant{
		Key:        "mystack-v1",
		StackKey:   "mystack",
		VariantDir: "v1",
		Files: []registry.FileMapping{
			{TemplatePath: "CLAUDE.md.tmpl", OutputPath: "CLAUDE.md"},
		},
	}

	dir := t.TempDir()
	// Create existing file
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("existing"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	vars := map[string]string{"ProjectName": "test"}
	results, err := eng.RenderAndWriteAll(variant, vars, dir, false)
	if err != nil {
		t.Fatalf("RenderAndWriteAll() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Written {
		t.Error("expected file to be skipped, but it was written")
	}
}

// TestAllTemplatesParse verifies that all templates in the real embed.FS parse successfully.
func TestAllTemplatesParse(t *testing.T) {
	eng := New(templates.FS, ".")
	if err := eng.ParseTemplates(); err != nil {
		t.Fatalf("ParseTemplates() on real templates failed: %v", err)
	}

	names := eng.TemplateNames()
	if len(names) == 0 {
		t.Fatal("expected at least one template, got none")
	}

	t.Logf("Successfully parsed %d templates: %v", len(names), names)
}
