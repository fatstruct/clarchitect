package registry

import (
	"strings"
	"unicode"
)

// Variable represents a template variable with its prompt, default, and flag info.
type Variable struct {
	Key          string // PascalCase key, e.g. "ProjectName"
	Prompt       string // Prompt text shown to user
	Default      string // Default value; empty string means required
	FlagOverride string // Optional override for the auto-derived flag name
}

// FlagName returns the CLI flag name for this variable.
// If FlagOverride is set, it is returned. Otherwise, the flag name
// is auto-derived from Key by converting PascalCase to kebab-case.
func (v Variable) FlagName() string {
	if v.FlagOverride != "" {
		return v.FlagOverride
	}
	return toKebabCase(v.Key)
}

// Required returns true if the variable has no default value and must be provided.
func (v Variable) Required() bool {
	return v.Default == ""
}

// FileMapping maps a template path to an output path.
type FileMapping struct {
	TemplatePath string // Relative to the templates dir within embed.FS
	OutputPath   string // Destination path relative to output base dir
}

// Variant represents a specific configuration variant (e.g., "go-chi").
type Variant struct {
	Key        string        // Unique key, e.g. "global", "go-chi"
	Label      string        // Human-readable label
	Vars       []Variable    // Variables required by this variant
	Files      []FileMapping // Template-to-output file mappings
	StackKey   string        // Back-reference to parent stack key
	VariantDir string        // Subdirectory within stack for templates (e.g., "chi" for go/chi/)
}

// Stack represents a language or platform grouping.
type Stack struct {
	Key      string    // Unique key, e.g. "global", "go"
	Label    string    // Human-readable label
	Variants []Variant // Available variants
}

// --- Builder pattern ---

// stacks is the global registry of all stacks.
var stacks []Stack

// StackBuilder builds a Stack.
type StackBuilder struct {
	stack Stack
}

// VariantBuilder builds a Variant within a Stack.
type VariantBuilder struct {
	variant Variant
	parent  *StackBuilder
}

// NewStack starts building a new stack with the given key.
func NewStack(key string) *StackBuilder {
	return &StackBuilder{
		stack: Stack{Key: key},
	}
}

// Label sets the human-readable label for the stack.
func (sb *StackBuilder) Label(label string) *StackBuilder {
	sb.stack.Label = label
	return sb
}

// Variant starts building a new variant within this stack.
// The variantDir is derived from the key by stripping the stack prefix (e.g., "go-chi" → "chi").
// If the key equals the stack key (e.g., global), variantDir is empty (templates live directly in stack dir).
func (sb *StackBuilder) Variant(key string) *VariantBuilder {
	var variantDir string
	prefix := sb.stack.Key + "-"
	if key == sb.stack.Key {
		variantDir = ""
	} else if len(key) > len(prefix) && key[:len(prefix)] == prefix {
		variantDir = key[len(prefix):]
	} else {
		variantDir = key
	}
	return &VariantBuilder{
		variant: Variant{
			Key:        key,
			StackKey:   sb.stack.Key,
			VariantDir: variantDir,
		},
		parent: sb,
	}
}

// Label sets the human-readable label for the variant.
func (vb *VariantBuilder) Label(label string) *VariantBuilder {
	vb.variant.Label = label
	return vb
}

// Variable adds a variable to the variant.
func (vb *VariantBuilder) Variable(key, prompt, defaultVal string) *VariantBuilder {
	vb.variant.Vars = append(vb.variant.Vars, Variable{
		Key:     key,
		Prompt:  prompt,
		Default: defaultVal,
	})
	return vb
}

// Flag sets an override flag name on the last added variable.
func (vb *VariantBuilder) Flag(flagName string) *VariantBuilder {
	if len(vb.variant.Vars) > 0 {
		vb.variant.Vars[len(vb.variant.Vars)-1].FlagOverride = flagName
	}
	return vb
}

// File adds a file mapping to the variant.
func (vb *VariantBuilder) File(templatePath, outputPath string) *VariantBuilder {
	vb.variant.Files = append(vb.variant.Files, FileMapping{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
	})
	return vb
}

// Done finishes building the variant, adds it to the parent stack, and returns
// the stack builder for further chaining.
func (vb *VariantBuilder) Done() *StackBuilder {
	vb.parent.stack.Variants = append(vb.parent.stack.Variants, vb.variant)
	return vb.parent
}

// Register finalizes the stack and adds it to the global registry.
func (sb *StackBuilder) Register() {
	stacks = append(stacks, sb.stack)
}

// --- Query functions ---

// AllStacks returns all registered stacks.
func AllStacks() []Stack {
	result := make([]Stack, len(stacks))
	copy(result, stacks)
	return result
}

// LookupVariant finds a variant by its key across all stacks.
func LookupVariant(key string) (Variant, bool) {
	for _, s := range stacks {
		for _, v := range s.Variants {
			if v.Key == key {
				return v, true
			}
		}
	}
	return Variant{}, false
}

// ResetRegistry clears all registered stacks. Used in tests.
func ResetRegistry() {
	stacks = nil
}

// --- Kebab-case conversion ---

// toKebabCase converts a PascalCase string to kebab-case.
// Examples:
//
//	ProjectName    → project-name
//	GoModule       → go-module
//	MinIOSVersion  → min-ios-version
//	BundleID       → bundle-id
func toKebabCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				// Look at what's around this uppercase letter to decide
				// whether to insert a hyphen.
				prevIsUpper := unicode.IsUpper(runes[i-1])
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if !prevIsUpper || nextIsLower {
					result.WriteRune('-')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// --- Default registrations ---

func init() {
	NewStack("global").
		Label("Global").
		Variant("global").
		Label("Global Config").
		Variable("AuthorName", "Your name", "").
		Variable("PreferredTestStyle", "Testing philosophy", "test alongside implementation").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		Done().
		Register()

	NewStack("typescript").
		Label("TypeScript").
		Variant("typescript-nextjs").
		Label("TypeScript + Next.js").
		Variable("ProjectName", "Project name", "").
		Variable("NodeVersion", "Node.js version", "22").
		Variable("PackageManager", "Package manager", "pnpm").
		Variable("NextVersion", "Next.js version", "15").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		File("rules/code-style.md.tmpl", ".claude/rules/code-style.md").
		File("rules/testing.md.tmpl", ".claude/rules/testing.md").
		Done().
		Variant("typescript-express").
		Label("TypeScript + Express API").
		Variable("ProjectName", "Project name", "").
		Variable("NodeVersion", "Node.js version", "22").
		Variable("PackageManager", "Package manager", "pnpm").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		File("rules/code-style.md.tmpl", ".claude/rules/code-style.md").
		File("rules/testing.md.tmpl", ".claude/rules/testing.md").
		Done().
		Register()

	NewStack("swift").
		Label("Swift").
		Variant("swift-swiftui").
		Label("Swift + SwiftUI (iOS)").
		Variable("ProjectName", "Project name", "").
		Variable("BundleID", "Bundle identifier", "").
		Variable("MinIOSVersion", "Minimum iOS version", "17").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		File("rules/swiftui.md.tmpl", ".claude/rules/swiftui.md").
		File("rules/testing.md.tmpl", ".claude/rules/testing.md").
		Done().
		Register()

	NewStack("go").
		Label("Go").
		Variant("go-chi").
		Label("Go + Chi Router").
		Variable("ProjectName", "Project name", "").
		Variable("GoModule", "Go module path", "").
		File("CLAUDE.md.tmpl", "CLAUDE.md").
		File("rules/api-conventions.md.tmpl", ".claude/rules/api-conventions.md").
		File("rules/testing.md.tmpl", ".claude/rules/testing.md").
		Done().
		Register()
}
