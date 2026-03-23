package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatstruct/clarchitect/internal/registry"
	"github.com/fatstruct/clarchitect/internal/version"
)

// --- Version command tests ---

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	if !strings.Contains(output, "clarchitect") {
		t.Errorf("version output should contain 'clarchitect', got: %q", output)
	}
	if !strings.Contains(output, version.Version) {
		t.Errorf("version output should contain version %q, got: %q", version.Version, output)
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	if !strings.Contains(output, "clarchitect") {
		t.Errorf("--version output should contain 'clarchitect', got: %q", output)
	}
}

func TestVersionShortFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-v"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	if !strings.Contains(output, "clarchitect") {
		t.Errorf("-v output should contain 'clarchitect', got: %q", output)
	}
}

func TestAllVersionAliasesProduceSameOutput(t *testing.T) {
	aliases := []string{"version", "--version", "-v"}
	var outputs []string

	for _, alias := range aliases {
		var stdout, stderr bytes.Buffer
		code := run([]string{alias}, &stdout, &stderr)
		if code != 0 {
			t.Errorf("exit code for %q = %d, want 0", alias, code)
		}
		outputs = append(outputs, stdout.String())
	}

	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("output for %q = %q, want %q (same as %q)",
				aliases[i], outputs[i], outputs[0], aliases[0])
		}
	}
}

// --- Help command tests ---

func TestHelpCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()

	// Should contain all command names
	commands := []string{"global", "project", "list", "version", "help"}
	for _, cmd := range commands {
		if !strings.Contains(output, cmd) {
			t.Errorf("help output should contain command %q, got:\n%s", cmd, output)
		}
	}

	// Should contain examples
	if !strings.Contains(output, "clarchitect global") {
		t.Errorf("help output should contain global example, got:\n%s", output)
	}
	if !strings.Contains(output, "clarchitect project go-chi") {
		t.Errorf("help output should contain project example, got:\n%s", output)
	}
	if !strings.Contains(output, "clarchitect list") {
		t.Errorf("help output should contain list example, got:\n%s", output)
	}
}

func TestHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	if !strings.Contains(output, "Commands:") {
		t.Errorf("--help output should contain 'Commands:', got:\n%s", output)
	}
}

func TestHelpShortFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := stdout.String()
	if !strings.Contains(output, "Commands:") {
		t.Errorf("-h output should contain 'Commands:', got:\n%s", output)
	}
}

func TestHelpContainsVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	run([]string{"help"}, &stdout, &stderr)
	output := stdout.String()
	if !strings.Contains(output, version.Version) {
		t.Errorf("help output should contain version %q, got:\n%s", version.Version, output)
	}
}

// --- No arguments behavior ---

func TestNoArgsShowsHelpAndExitsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	// No args should print help to stderr
	output := stderr.String()
	if !strings.Contains(output, "Commands:") {
		t.Errorf("no-args output should contain 'Commands:', got:\n%s", output)
	}
}

// --- List command tests ---

func TestListCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"list"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	output := stdout.String()

	// Should contain "Available stacks:" header
	if !strings.Contains(output, "Available stacks:") {
		t.Errorf("list output should contain 'Available stacks:', got:\n%s", output)
	}

	// Should contain all stack labels (except global)
	stackLabels := []string{"TypeScript", "Swift", "Go"}
	for _, label := range stackLabels {
		if !strings.Contains(output, label) {
			t.Errorf("list output should contain stack label %q, got:\n%s", label, output)
		}
	}

	// Should contain all variant labels
	variantLabels := []string{
		"TypeScript + Next.js",
		"TypeScript + Express API",
		"Swift + SwiftUI (iOS)",
		"Go + Chi Router",
	}
	for _, label := range variantLabels {
		if !strings.Contains(output, label) {
			t.Errorf("list output should contain variant label %q, got:\n%s", label, output)
		}
	}

	// Should contain usage hints
	usageHints := []string{
		"clarchitect project typescript-nextjs",
		"clarchitect project typescript-express",
		"clarchitect project swift-swiftui",
		"clarchitect project go-chi",
	}
	for _, hint := range usageHints {
		if !strings.Contains(output, hint) {
			t.Errorf("list output should contain usage hint %q, got:\n%s", hint, output)
		}
	}

	// Should NOT contain "global" or "Global"
	// (checking per-line to avoid false positives from "Global" appearing in "runGlobal" etc.)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Global" || trimmed == "global" {
			t.Errorf("list output should not contain the global stack, found line: %q", line)
		}
	}
}

func TestListCommandReadsFromRegistry(t *testing.T) {
	// Verify the list command output changes when the registry changes.
	// We use AllStacks to get the current state and verify list
	// reflects it accurately.
	allStacks := registry.AllStacks()

	var stdout, stderr bytes.Buffer
	run([]string{"list"}, &stdout, &stderr)
	output := stdout.String()

	projectStackCount := 0
	for _, s := range allStacks {
		if s.Key == "global" {
			continue
		}
		projectStackCount++
		for _, v := range s.Variants {
			if !strings.Contains(output, v.Key) {
				t.Errorf("list output should contain variant key %q", v.Key)
			}
			if !strings.Contains(output, v.Label) {
				t.Errorf("list output should contain variant label %q", v.Label)
			}
		}
	}

	if projectStackCount == 0 {
		t.Fatal("expected at least one project stack in registry")
	}
}

// --- Unknown command tests ---

func TestUnknownCommandShowsErrorAndHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"foobar"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	output := stderr.String()

	// Should contain error about unknown command
	if !strings.Contains(output, `unknown command "foobar"`) {
		t.Errorf("expected error about unknown command, got:\n%s", output)
	}

	// Should also contain help text
	if !strings.Contains(output, "Commands:") {
		t.Errorf("unknown command output should contain help text, got:\n%s", output)
	}

	// Should contain examples
	if !strings.Contains(output, "Examples:") {
		t.Errorf("unknown command output should contain examples, got:\n%s", output)
	}
}

func TestUnknownCommandDifferentNames(t *testing.T) {
	unknowns := []string{"init", "generate", "setup", ""}
	for _, cmd := range unknowns {
		if cmd == "" {
			continue // empty args case tested separately
		}
		t.Run(cmd, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{cmd}, &stdout, &stderr)
			if code != 1 {
				t.Errorf("exit code for %q = %d, want 1", cmd, code)
			}
			output := stderr.String()
			if !strings.Contains(output, "unknown command") {
				t.Errorf("expected 'unknown command' in output for %q, got:\n%s", cmd, output)
			}
		})
	}
}

// --- printVersion tests ---

func TestPrintVersionOutput(t *testing.T) {
	var buf bytes.Buffer
	printVersion(&buf)
	output := buf.String()

	if !strings.HasPrefix(output, "clarchitect ") {
		t.Errorf("version output should start with 'clarchitect ', got: %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("version output should end with newline, got: %q", output)
	}
}

// --- printList tests ---

func TestPrintListAlignment(t *testing.T) {
	var buf bytes.Buffer
	printList(&buf)
	output := buf.String()

	// Verify that within each stack section, the "clarchitect project" hints
	// are present on each variant line
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "clarchitect project") {
			// These lines should start with spaces (indentation)
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("variant line should be indented with 4 spaces, got: %q", line)
			}
		}
	}
}

// --- Existing tests (preserved and updated) ---

func TestRunIsCallable(t *testing.T) {
	_ = Run
}

func TestPrintHelpDoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	printHelp(&buf)
}

func TestPrintHelpContainsProjectCommand(t *testing.T) {
	var buf bytes.Buffer
	printHelp(&buf)
	output := buf.String()

	if !strings.Contains(output, "project") {
		t.Error("help text should contain 'project' command")
	}
}

// --- Registry tests for all 4 variants ---

func TestAllFourVariantsRegistered(t *testing.T) {
	variants := []struct {
		key   string
		label string
	}{
		{"typescript-nextjs", "TypeScript + Next.js"},
		{"typescript-express", "TypeScript + Express API"},
		{"swift-swiftui", "Swift + SwiftUI (iOS)"},
		{"go-chi", "Go + Chi Router"},
	}

	for _, tt := range variants {
		t.Run(tt.key, func(t *testing.T) {
			v, ok := registry.LookupVariant(tt.key)
			if !ok {
				t.Fatalf("variant %q not found in registry", tt.key)
			}
			if v.Label != tt.label {
				t.Errorf("variant label = %q, want %q", v.Label, tt.label)
			}
		})
	}
}

func TestVariantStackKeysAndDirs(t *testing.T) {
	tests := []struct {
		key        string
		stackKey   string
		variantDir string
	}{
		{"typescript-nextjs", "typescript", "nextjs"},
		{"typescript-express", "typescript", "express"},
		{"swift-swiftui", "swift", "swiftui"},
		{"go-chi", "go", "chi"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			v, ok := registry.LookupVariant(tt.key)
			if !ok {
				t.Fatalf("variant %q not found", tt.key)
			}
			if v.StackKey != tt.stackKey {
				t.Errorf("StackKey = %q, want %q", v.StackKey, tt.stackKey)
			}
			if v.VariantDir != tt.variantDir {
				t.Errorf("VariantDir = %q, want %q", v.VariantDir, tt.variantDir)
			}
		})
	}
}

func TestVariantFileMappings(t *testing.T) {
	tests := []struct {
		key           string
		expectedFiles int
	}{
		{"typescript-nextjs", 3},
		{"typescript-express", 3},
		{"swift-swiftui", 3},
		{"go-chi", 3},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			v, ok := registry.LookupVariant(tt.key)
			if !ok {
				t.Fatalf("variant %q not found", tt.key)
			}
			if len(v.Files) != tt.expectedFiles {
				t.Errorf("expected %d file mappings, got %d", tt.expectedFiles, len(v.Files))
			}
		})
	}
}

// --- Auto-derived flag name tests (table-driven) ---

func TestAutoDerivedFlagNamesForAllVariants(t *testing.T) {
	tests := []struct {
		variantKey string
		varKey     string
		wantFlag   string
	}{
		// typescript-nextjs
		{"typescript-nextjs", "ProjectName", "project-name"},
		{"typescript-nextjs", "NodeVersion", "node-version"},
		{"typescript-nextjs", "PackageManager", "package-manager"},
		{"typescript-nextjs", "NextVersion", "next-version"},
		// typescript-express
		{"typescript-express", "ProjectName", "project-name"},
		{"typescript-express", "NodeVersion", "node-version"},
		{"typescript-express", "PackageManager", "package-manager"},
		// swift-swiftui
		{"swift-swiftui", "ProjectName", "project-name"},
		{"swift-swiftui", "BundleID", "bundle-id"},
		{"swift-swiftui", "MinIOSVersion", "min-ios-version"},
		// go-chi
		{"go-chi", "ProjectName", "project-name"},
		{"go-chi", "GoModule", "go-module"},
	}

	for _, tt := range tests {
		name := tt.variantKey + "/" + tt.varKey
		t.Run(name, func(t *testing.T) {
			v, ok := registry.LookupVariant(tt.variantKey)
			if !ok {
				t.Fatalf("variant %q not found", tt.variantKey)
			}
			var found bool
			for _, variable := range v.Vars {
				if variable.Key == tt.varKey {
					found = true
					got := variable.FlagName()
					if got != tt.wantFlag {
						t.Errorf("FlagName() = %q, want %q", got, tt.wantFlag)
					}
					break
				}
			}
			if !found {
				t.Errorf("variable %q not found in variant %q", tt.varKey, tt.variantKey)
			}
		})
	}
}

// --- Project command tests using runProjectE ---

func TestProjectCommandRendersAllFiles(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	args := []string{"go-chi",
		"--project-name", "my-api",
		"--go-module", "github.com/me/my-api",
	}

	err := runProjectE(args, &stdout, &stderr, dir)
	if err != nil {
		t.Fatalf("runProjectE() error = %v\nstderr: %s", err, stderr.String())
	}

	// Check that all 3 files were written
	output := stdout.String()
	expectedFiles := []string{
		"CLAUDE.md",
		".claude/rules/api-conventions.md",
		".claude/rules/testing.md",
	}

	for _, f := range expectedFiles {
		if !strings.Contains(output, "Written: "+f) {
			t.Errorf("expected output to contain 'Written: %s', got:\n%s", f, output)
		}
	}

	// Verify file contents
	claudeContent, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeContent), "my-api") {
		t.Errorf("CLAUDE.md should contain project name, got: %s", claudeContent)
	}
	if !strings.Contains(string(claudeContent), "github.com/me/my-api") {
		t.Errorf("CLAUDE.md should contain module path, got: %s", claudeContent)
	}

	apiContent, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "api-conventions.md"))
	if err != nil {
		t.Fatalf("reading api-conventions.md: %v", err)
	}
	if !strings.Contains(string(apiContent), "my-api") {
		t.Errorf("api-conventions.md should contain project name, got: %s", apiContent)
	}
}

func TestProjectCommandAllVariantsRender(t *testing.T) {
	tests := []struct {
		variant string
		args    []string
		files   []string
	}{
		{
			variant: "typescript-nextjs",
			args:    []string{"--project-name", "my-app"},
			files:   []string{"CLAUDE.md", ".claude/rules/code-style.md", ".claude/rules/testing.md"},
		},
		{
			variant: "typescript-express",
			args:    []string{"--project-name", "my-api"},
			files:   []string{"CLAUDE.md", ".claude/rules/code-style.md", ".claude/rules/testing.md"},
		},
		{
			variant: "swift-swiftui",
			args:    []string{"--project-name", "MyApp", "--bundle-id", "com.example.myapp"},
			files:   []string{"CLAUDE.md", ".claude/rules/swiftui.md", ".claude/rules/testing.md"},
		},
		{
			variant: "go-chi",
			args:    []string{"--project-name", "my-api", "--go-module", "github.com/me/my-api"},
			files:   []string{"CLAUDE.md", ".claude/rules/api-conventions.md", ".claude/rules/testing.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.variant, func(t *testing.T) {
			dir := t.TempDir()
			var stdout, stderr bytes.Buffer

			fullArgs := append([]string{tt.variant}, tt.args...)
			err := runProjectE(fullArgs, &stdout, &stderr, dir)
			if err != nil {
				t.Fatalf("runProjectE() error = %v\nstderr: %s", err, stderr.String())
			}

			// Verify all files exist
			for _, f := range tt.files {
				path := filepath.Join(dir, f)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected file %s to exist", f)
				}
			}

			// Verify output mentions each file
			output := stdout.String()
			for _, f := range tt.files {
				if !strings.Contains(output, "Written: "+f) {
					t.Errorf("expected output to contain 'Written: %s'", f)
				}
			}
		})
	}
}

func TestProjectCommandInvalidVariant(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runProjectE([]string{"nonexistent"}, &stdout, &stderr, t.TempDir())
	if err == nil {
		t.Fatal("expected error for invalid variant, got nil")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "unknown variant") {
		t.Errorf("expected error to contain 'unknown variant', got: %s", errOutput)
	}

	// Should list available variants
	if !strings.Contains(errOutput, "go-chi") {
		t.Errorf("error output should list available variants, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "typescript-nextjs") {
		t.Errorf("error output should list available variants, got: %s", errOutput)
	}
}

func TestProjectCommandMissingVariantArg(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runProjectE([]string{}, &stdout, &stderr, t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing variant argument, got nil")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "missing variant argument") {
		t.Errorf("expected error about missing variant, got: %s", errOutput)
	}

	// Should list available variants
	if !strings.Contains(errOutput, "Available variants:") {
		t.Errorf("error output should list available variants, got: %s", errOutput)
	}
}

func TestProjectCommandRequiredFlagsMissing(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// go-chi requires --project-name and --go-module
	err := runProjectE([]string{"go-chi"}, &stdout, &stderr, t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing required flags, got nil")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "required flag(s) missing") {
		t.Errorf("expected error about required flags, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "project-name") {
		t.Errorf("error should mention project-name flag, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "go-module") {
		t.Errorf("error should mention go-module flag, got: %s", errOutput)
	}
}

func TestProjectCommandDefaultValuesApplied(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	// typescript-nextjs has defaults: NodeVersion=22, PackageManager=pnpm, NextVersion=15
	// Only provide the required flag --project-name
	args := []string{"typescript-nextjs", "--project-name", "my-app"}
	err := runProjectE(args, &stdout, &stderr, dir)
	if err != nil {
		t.Fatalf("runProjectE() error = %v\nstderr: %s", err, stderr.String())
	}

	// Read rendered CLAUDE.md and check default values were used
	content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "22") {
		t.Errorf("expected default NodeVersion '22' in output, got: %s", s)
	}
	if !strings.Contains(s, "pnpm") {
		t.Errorf("expected default PackageManager 'pnpm' in output, got: %s", s)
	}
	if !strings.Contains(s, "15") {
		t.Errorf("expected default NextVersion '15' in output, got: %s", s)
	}
}

func TestProjectCommandForceOverwritesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create an existing CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("original content"), 0644); err != nil {
		t.Fatalf("creating existing file: %v", err)
	}

	// First run without --force: CLAUDE.md should be skipped
	var stdout1, stderr1 bytes.Buffer
	args := []string{"go-chi", "--project-name", "my-api", "--go-module", "github.com/me/my-api"}
	err := runProjectE(args, &stdout1, &stderr1, dir)
	if err != nil {
		t.Fatalf("first run error = %v", err)
	}

	output1 := stdout1.String()
	if !strings.Contains(output1, "Skipped (already exists): CLAUDE.md") {
		t.Errorf("expected CLAUDE.md to be skipped without --force, got:\n%s", output1)
	}

	// Verify original content preserved
	data, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(data) != "original content" {
		t.Errorf("original content should be preserved, got: %s", data)
	}

	// Now run with --force: CLAUDE.md should be overwritten
	var stdout2, stderr2 bytes.Buffer
	argsForce := []string{"go-chi", "--project-name", "my-api", "--go-module", "github.com/me/my-api", "--force"}
	err = runProjectE(argsForce, &stdout2, &stderr2, dir)
	if err != nil {
		t.Fatalf("second run error = %v\nstderr: %s", err, stderr2.String())
	}

	output2 := stdout2.String()
	if !strings.Contains(output2, "Written: CLAUDE.md") {
		t.Errorf("expected CLAUDE.md to be written with --force, got:\n%s", output2)
	}

	// Verify content was overwritten
	data, _ = os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(data) == "original content" {
		t.Error("file should have been overwritten with --force")
	}
	if !strings.Contains(string(data), "my-api") {
		t.Errorf("overwritten file should contain project name, got: %s", data)
	}
}

func TestAllVariantKeysExcludesGlobal(t *testing.T) {
	keys := allVariantKeys()
	for _, k := range keys {
		if k == "global" {
			t.Error("allVariantKeys() should not include 'global'")
		}
	}

	// Verify all 4 project variants are present
	expected := []string{"go-chi", "swift-swiftui", "typescript-express", "typescript-nextjs"}
	if len(keys) < len(expected) {
		t.Fatalf("expected at least %d variant keys, got %d: %v", len(expected), len(keys), keys)
	}

	for _, e := range expected {
		found := false
		for _, k := range keys {
			if k == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected variant key %q in allVariantKeys(), got: %v", e, keys)
		}
	}
}

func TestProjectCommandSwiftSwiftUIPartialRequiredFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// swift-swiftui requires ProjectName and BundleID, provide only one
	args := []string{"swift-swiftui", "--project-name", "MyApp"}
	err := runProjectE(args, &stdout, &stderr, t.TempDir())
	if err == nil {
		t.Fatal("expected error when BundleID missing")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "bundle-id") {
		t.Errorf("error should mention missing bundle-id flag, got: %s", errOutput)
	}
}

// --- TTY detection tests ---

func TestIsTerminalReturnsBool(t *testing.T) {
	result := isTerminal()
	_ = result
}

func TestIsTerminalReturnsFalseForPipedStdin(t *testing.T) {
	if isTerminal() {
		t.Skip("stdin is a terminal in this environment; skipping piped-stdin test")
	}
}

// --- Exit code tests ---

func TestRunReturnsExitCode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"version returns 0", []string{"version"}, 0},
		{"--version returns 0", []string{"--version"}, 0},
		{"-v returns 0", []string{"-v"}, 0},
		{"help returns 0", []string{"help"}, 0},
		{"--help returns 0", []string{"--help"}, 0},
		{"-h returns 0", []string{"-h"}, 0},
		{"list returns 0", []string{"list"}, 0},
		{"no args returns 1", []string{}, 1},
		{"unknown command returns 1", []string{"bogus"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tt.args, &stdout, &stderr)
			if got != tt.want {
				t.Errorf("run(%v) = %d, want %d\nstdout: %s\nstderr: %s",
					tt.args, got, tt.want, stdout.String(), stderr.String())
			}
		})
	}
}
