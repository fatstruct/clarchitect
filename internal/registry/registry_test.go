package registry

import (
	"testing"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ProjectName", "project-name"},
		{"GoModule", "go-module"},
		{"MinIOSVersion", "min-ios-version"},
		{"BundleID", "bundle-id"},
		{"AuthorName", "author-name"},
		{"PreferredTestStyle", "preferred-test-style"},
		{"", ""},
		{"Name", "name"},
		{"X", "x"},
		{"HTTPServer", "http-server"},
		{"NodeVersion", "node-version"},
		{"PackageManager", "package-manager"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuilderCreatesStack(t *testing.T) {
	ResetRegistry()

	NewStack("teststack").
		Label("Test Stack").
		Variant("teststack-v1").
		Label("Test Variant One").
		Variable("ProjectName", "Project name", "").
		Variable("Version", "Version number", "1.0").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		Done().
		Register()

	stacks := AllStacks()
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}

	s := stacks[0]
	if s.Key != "teststack" {
		t.Errorf("stack key = %q, want %q", s.Key, "teststack")
	}
	if s.Label != "Test Stack" {
		t.Errorf("stack label = %q, want %q", s.Label, "Test Stack")
	}
	if len(s.Variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(s.Variants))
	}

	v := s.Variants[0]
	if v.Key != "teststack-v1" {
		t.Errorf("variant key = %q, want %q", v.Key, "teststack-v1")
	}
	if v.Label != "Test Variant One" {
		t.Errorf("variant label = %q, want %q", v.Label, "Test Variant One")
	}
	if v.StackKey != "teststack" {
		t.Errorf("variant stack key = %q, want %q", v.StackKey, "teststack")
	}
	if len(v.Vars) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(v.Vars))
	}

	// Check first variable (required)
	if v.Vars[0].Key != "ProjectName" {
		t.Errorf("var[0].Key = %q, want %q", v.Vars[0].Key, "ProjectName")
	}
	if v.Vars[0].Prompt != "Project name" {
		t.Errorf("var[0].Prompt = %q, want %q", v.Vars[0].Prompt, "Project name")
	}
	if !v.Vars[0].Required() {
		t.Error("var[0] should be required (empty default)")
	}

	// Check second variable (has default)
	if v.Vars[1].Key != "Version" {
		t.Errorf("var[1].Key = %q, want %q", v.Vars[1].Key, "Version")
	}
	if v.Vars[1].Default != "1.0" {
		t.Errorf("var[1].Default = %q, want %q", v.Vars[1].Default, "1.0")
	}
	if v.Vars[1].Required() {
		t.Error("var[1] should not be required (has default)")
	}

	// Check file mapping
	if len(v.Files) != 1 {
		t.Fatalf("expected 1 file mapping, got %d", len(v.Files))
	}
	if v.Files[0].TemplatePath != "CLAUDE.md.tmpl" {
		t.Errorf("file template path = %q, want %q", v.Files[0].TemplatePath, "CLAUDE.md.tmpl")
	}
	if v.Files[0].OutputPath != "CLAUDE.md" {
		t.Errorf("file output path = %q, want %q", v.Files[0].OutputPath, "CLAUDE.md")
	}
}

func TestVariableFlagNameDerivation(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"ProjectName", "project-name"},
		{"GoModule", "go-module"},
		{"MinIOSVersion", "min-ios-version"},
		{"BundleID", "bundle-id"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			v := Variable{Key: tt.key}
			if got := v.FlagName(); got != tt.want {
				t.Errorf("Variable{Key: %q}.FlagName() = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestVariableFlagNameOverride(t *testing.T) {
	v := Variable{Key: "ProjectName", FlagOverride: "custom-name"}
	if got := v.FlagName(); got != "custom-name" {
		t.Errorf("FlagName() with override = %q, want %q", got, "custom-name")
	}
}

func TestFlagOverrideViaBuilder(t *testing.T) {
	ResetRegistry()

	NewStack("test").
		Label("Test").
		Variant("test-v1").
		Label("Test V1").
		Variable("ProjectName", "Project name", "").
		Flag("custom-project").
		Done().
		Register()

	v, ok := LookupVariant("test-v1")
	if !ok {
		t.Fatal("variant not found")
	}

	if len(v.Vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(v.Vars))
	}

	if got := v.Vars[0].FlagName(); got != "custom-project" {
		t.Errorf("FlagName() = %q, want %q", got, "custom-project")
	}
}

func TestLookupVariant(t *testing.T) {
	ResetRegistry()

	NewStack("s1").
		Label("Stack One").
		Variant("s1-a").
		Label("Variant A").
		Variable("Name", "Name", "").
		File("a.tmpl", "a.md").
		Done().
		Variant("s1-b").
		Label("Variant B").
		Variable("Name", "Name", "default").
		File("b.tmpl", "b.md").
		Done().
		Register()

	// Lookup existing variant
	v, ok := LookupVariant("s1-a")
	if !ok {
		t.Fatal("expected to find variant s1-a")
	}
	if v.Label != "Variant A" {
		t.Errorf("label = %q, want %q", v.Label, "Variant A")
	}

	v, ok = LookupVariant("s1-b")
	if !ok {
		t.Fatal("expected to find variant s1-b")
	}
	if v.Label != "Variant B" {
		t.Errorf("label = %q, want %q", v.Label, "Variant B")
	}

	// Lookup non-existing variant
	_, ok = LookupVariant("nonexistent")
	if ok {
		t.Error("expected not to find variant 'nonexistent'")
	}
}

func TestMultipleVariantsInStack(t *testing.T) {
	ResetRegistry()

	NewStack("multi").
		Label("Multi").
		Variant("multi-a").
		Label("A").
		Done().
		Variant("multi-b").
		Label("B").
		Done().
		Register()

	stacks := AllStacks()
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if len(stacks[0].Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(stacks[0].Variants))
	}
	if stacks[0].Variants[0].Key != "multi-a" {
		t.Errorf("variant[0].Key = %q, want %q", stacks[0].Variants[0].Key, "multi-a")
	}
	if stacks[0].Variants[1].Key != "multi-b" {
		t.Errorf("variant[1].Key = %q, want %q", stacks[0].Variants[1].Key, "multi-b")
	}
}
