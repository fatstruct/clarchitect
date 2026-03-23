package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatstruct/clarchitect/internal/registry"
)

// ErrFileExists is returned when a file exists and force is false.
var ErrFileExists = errors.New("file already exists")

// Engine handles template parsing, rendering, and file writing.
type Engine struct {
	fsys      fs.FS
	basePath  string
	templates map[string]*template.Template
}

// New creates a new Engine with the given embedded filesystem and base path.
func New(fsys fs.FS, basePath string) *Engine {
	return &Engine{
		fsys:      fsys,
		basePath:  basePath,
		templates: make(map[string]*template.Template),
	}
}

// ParseTemplates walks the embedded filesystem and parses all .tmpl files.
// Returns an error if any template fails to parse.
func (e *Engine) ParseTemplates() error {
	return fs.WalkDir(e.fsys, e.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		data, err := fs.ReadFile(e.fsys, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		// Store with a key relative to basePath for easier lookup
		relPath, err := filepath.Rel(e.basePath, path)
		if err != nil {
			relPath = path
		}

		tmpl, err := template.New(relPath).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		e.templates[relPath] = tmpl
		return nil
	})
}

// Render renders a single template with the given variables.
// templatePath is relative to the base path (e.g., "global/CLAUDE.md.tmpl").
func (e *Engine) Render(templatePath string, vars map[string]string) (string, error) {
	tmpl, ok := e.templates[templatePath]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templatePath)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("rendering template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// WriteFile writes content to the given path. If the file exists and force is
// false, it returns ErrFileExists. If force is true, existing files are overwritten.
// Parent directories are created with 0755 permissions.
func (e *Engine) WriteFile(path string, content string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return ErrFileExists
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", path, err)
	}

	return nil
}

// FileResult represents the outcome of rendering and writing a single file.
type FileResult struct {
	OutputPath string
	Written    bool   // true if file was written, false if skipped
	Error      error  // non-nil if an error occurred (other than skip)
}

// RenderAndWriteAll renders all files for a variant and writes them to baseDir.
// Returns a slice of results indicating what happened with each file.
func (e *Engine) RenderAndWriteAll(variant registry.Variant, vars map[string]string, baseDir string, force bool) ([]FileResult, error) {
	var results []FileResult

	for _, fm := range variant.Files {
		// Build the full template path: <stackKey>/<variantDir>/<templatePath>
		templatePath := filepath.Join(variant.StackKey, variant.VariantDir, fm.TemplatePath)

		content, err := e.Render(templatePath, vars)
		if err != nil {
			return results, fmt.Errorf("rendering %s: %w", fm.TemplatePath, err)
		}

		outputPath := filepath.Join(baseDir, fm.OutputPath)
		writeErr := e.WriteFile(outputPath, content, force)

		result := FileResult{OutputPath: fm.OutputPath}
		if writeErr != nil {
			if errors.Is(writeErr, ErrFileExists) {
				result.Written = false
			} else {
				result.Error = writeErr
				results = append(results, result)
				return results, writeErr
			}
		} else {
			result.Written = true
		}

		results = append(results, result)
	}

	return results, nil
}

// TemplateNames returns the names of all parsed templates.
func (e *Engine) TemplateNames() []string {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}
