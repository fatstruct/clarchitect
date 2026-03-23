package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/fatstruct/clarchitect/internal/engine"
	"github.com/fatstruct/clarchitect/internal/registry"
	"github.com/fatstruct/clarchitect/internal/tui"
	"github.com/fatstruct/clarchitect/internal/version"
	"github.com/fatstruct/clarchitect/templates"
)

// Run is the main entry point for the CLI. It dispatches commands based on args
// and returns an exit code. The caller should pass the result to os.Exit.
func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

// run is the internal implementation that accepts writers for testability.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stderr)
		return 1
	}

	switch args[0] {
	case "version", "--version", "-v":
		printVersion(stdout)
		return 0
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	case "list":
		printList(stdout)
		return 0
	case "global":
		return runGlobal(args[1:], stdout, stderr)
	case "project":
		if err := runProjectE(args[1:], stdout, stderr, ""); err != nil {
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n", args[0])
		printHelp(stderr)
		return 1
	}
}

// printVersion prints the version string.
func printVersion(w io.Writer) {
	fmt.Fprintf(w, "clarchitect %s\n", version.Version)
}

// printList reads all stacks and variants from the registry and prints
// a formatted tree to the given writer.
func printList(w io.Writer) {
	allStacks := registry.AllStacks()

	fmt.Fprintln(w, "Available stacks:")

	for _, stack := range allStacks {
		// Skip the global stack — it is not a project stack
		if stack.Key == "global" {
			continue
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %s\n", stack.Label)

		// Find the longest label for alignment
		maxLabelLen := 0
		for _, v := range stack.Variants {
			if len(v.Label) > maxLabelLen {
				maxLabelLen = len(v.Label)
			}
		}

		for _, v := range stack.Variants {
			padding := strings.Repeat(" ", maxLabelLen-len(v.Label)+4)
			fmt.Fprintf(w, "    %s%sclarchitect project %s\n", v.Label, padding, v.Key)
		}
	}

	fmt.Fprintln(w)
}

// printHelp prints the full help text with usage synopsis, commands, and examples.
func printHelp(w io.Writer) {
	fmt.Fprintf(w, "clarchitect %s\n", version.Version)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Scaffold Claude Code configuration files from curated templates.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage: clarchitect <command> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  global     Set up ~/.claude/CLAUDE.md with your coding identity")
	fmt.Fprintln(w, "  project    Generate project-specific Claude rules from a stack template")
	fmt.Fprintln(w, "  list       Show all available stacks and variants")
	fmt.Fprintln(w, "  version    Print the version")
	fmt.Fprintln(w, "  help       Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, `  clarchitect global --author-name "Your Name"`)
	fmt.Fprintln(w, `  clarchitect project go-chi --project-name my-api --go-module github.com/me/my-api`)
	fmt.Fprintln(w, `  clarchitect list`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run 'clarchitect <command> --help' for more information on a command.")
}

// isTerminal reports whether stdin is connected to a terminal.
func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func runGlobal(args []string, stdout, stderr io.Writer) int {
	variant, ok := registry.LookupVariant("global")
	if !ok {
		fmt.Fprintln(stderr, "Error: global variant not found in registry")
		return 1
	}

	// Check if any flags were provided. If stdin is a terminal and no flags
	// are passed, launch the interactive TUI.
	hasFlags := false
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			hasFlags = true
			break
		}
	}

	if isTerminal() && !hasFlags {
		return runGlobalInteractive(variant, stderr)
	}

	// Set up flags from registry variables
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	fs.SetOutput(stderr)
	flagValues := make(map[string]*string)
	for _, v := range variant.Vars {
		val := fs.String(v.FlagName(), v.Default, v.Prompt)
		flagValues[v.Key] = val
	}

	forceFlag := fs.Bool("force", false, "Overwrite existing files")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Collect variable values and check required ones
	vars := make(map[string]string)
	for _, v := range variant.Vars {
		val := *flagValues[v.Key]
		if val == "" && v.Required() {
			fmt.Fprintf(stderr, "Error: required flag --%s is missing\n", v.FlagName())
			return 1
		}
		vars[v.Key] = val
	}

	// Resolve output directory: ~/.claude/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "Error: could not determine home directory: %v\n", err)
		return 1
	}
	outputDir := filepath.Join(homeDir, ".claude")

	// Initialize engine and parse templates
	eng := engine.New(templates.FS, ".")
	if err := eng.ParseTemplates(); err != nil {
		fmt.Fprintf(stderr, "Error: failed to parse templates: %v\n", err)
		return 1
	}

	// Render and write all files
	results, err := eng.RenderAndWriteAll(variant, vars, outputDir, *forceFlag)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	// Print results
	for _, r := range results {
		fullPath := filepath.Join(outputDir, r.OutputPath)
		if r.Written {
			fmt.Fprintf(stdout, "Written: %s\n", fullPath)
		} else {
			fmt.Fprintf(stdout, "Skipped (already exists): %s\n", fullPath)
		}
	}

	return 0
}

// runGlobalInteractive launches the Bubbletea TUI for the global command.
func runGlobalInteractive(variant registry.Variant, stderr io.Writer) int {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "Error: could not determine home directory: %v\n", err)
		return 1
	}
	outputDir := filepath.Join(homeDir, ".claude")

	eng := engine.New(templates.FS, ".")
	if err := eng.ParseTemplates(); err != nil {
		fmt.Fprintf(stderr, "Error: failed to parse templates: %v\n", err)
		return 1
	}

	model := tui.NewGlobalForm(variant, eng, outputDir)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// allVariantKeys returns a sorted list of all registered variant keys,
// excluding the "global" variant.
func allVariantKeys() []string {
	var keys []string
	for _, s := range registry.AllStacks() {
		for _, v := range s.Variants {
			if v.Key == "global" {
				continue
			}
			keys = append(keys, v.Key)
		}
	}
	sort.Strings(keys)
	return keys
}

// runProjectE is the testable core of the project command.
// It writes output to stdout and errors to stderr. If outputDir is empty,
// the current working directory is used. Returns a non-nil error on failure
// (after printing the error message to stderr).
func runProjectE(args []string, stdout, stderr io.Writer, outputDir string) error {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Error: missing variant argument")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Usage: clarchitect project <variant> [flags]")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Available variants:")
		for _, k := range allVariantKeys() {
			fmt.Fprintf(stderr, "  %s\n", k)
		}
		return fmt.Errorf("missing variant argument")
	}

	variantKey := args[0]
	variant, ok := registry.LookupVariant(variantKey)
	if !ok {
		fmt.Fprintf(stderr, "Error: unknown variant %q\n", variantKey)
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Available variants:")
		for _, k := range allVariantKeys() {
			fmt.Fprintf(stderr, "  %s\n", k)
		}
		return fmt.Errorf("unknown variant %q", variantKey)
	}

	// Set up flags from registry variables
	fs := flag.NewFlagSet("project "+variantKey, flag.ContinueOnError)
	fs.SetOutput(stderr)
	flagValues := make(map[string]*string)
	for _, v := range variant.Vars {
		val := fs.String(v.FlagName(), v.Default, v.Prompt)
		flagValues[v.Key] = val
	}

	forceFlag := fs.Bool("force", false, "Overwrite existing files")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// Collect variable values and check required ones
	vars := make(map[string]string)
	var missing []string
	for _, v := range variant.Vars {
		val := *flagValues[v.Key]
		if val == "" && v.Required() {
			missing = append(missing, v.FlagName())
		}
		vars[v.Key] = val
	}
	if len(missing) > 0 {
		msg := fmt.Sprintf("Error: required flag(s) missing: --%s", strings.Join(missing, ", --"))
		fmt.Fprintln(stderr, msg)
		return fmt.Errorf("required flags missing: %s", strings.Join(missing, ", "))
	}

	// Resolve output directory
	if outputDir == "" {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "Error: could not determine working directory: %v\n", err)
			return err
		}
	}

	// Initialize engine and parse templates
	eng := engine.New(templates.FS, ".")
	if err := eng.ParseTemplates(); err != nil {
		fmt.Fprintf(stderr, "Error: failed to parse templates: %v\n", err)
		return err
	}

	// Render and write all files
	results, err := eng.RenderAndWriteAll(variant, vars, outputDir, *forceFlag)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return err
	}

	// Print results
	for _, r := range results {
		if r.Written {
			fmt.Fprintf(stdout, "Written: %s\n", r.OutputPath)
		} else {
			fmt.Fprintf(stdout, "Skipped (already exists): %s\n", r.OutputPath)
		}
	}

	return nil
}
