package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/fatstruct/clarchitect/internal/engine"
	"github.com/fatstruct/clarchitect/internal/registry"
)

// projectPhase tracks the current phase of the project form.
type projectPhase int

const (
	projectPhaseSelectStack   projectPhase = iota // Choose a stack (e.g., Go, TypeScript)
	projectPhaseSelectVariant                     // Choose a variant within the stack
	projectPhaseCollectVars                       // Collect template variables
	projectPhaseProcessFiles                      // Render and write files
	projectPhaseConfirmOverwrite                  // Ask about overwriting an existing file
	projectPhaseDone                              // All done
)

// ProjectFormModel orchestrates the interactive project flow:
// stack selection -> variant selection -> variable collection -> file rendering.
type ProjectFormModel struct {
	stacks []registry.Stack // Project stacks (excludes "global")
	eng    *engine.Engine
	outputDir string

	phase projectPhase

	// Stack selection
	stackSelection SelectionModel
	selectedStack  int

	// Variant selection
	variantSelection SelectionModel
	selectedVariant  int

	// The resolved variant (set after selection or pre-provided)
	variant registry.Variant

	// Variable collection (same pattern as GlobalFormModel)
	varIndex   int
	textInput  TextInputModel
	vars       map[string]string
	varHistory []string

	// File processing
	fileIndex    int
	overwriteAll bool
	confirmModel ConfirmModel
	outcomes     []fileOutcome

	// Rendered content cache
	renderedContent map[int]string

	// Selection history (rendered views of completed selections)
	selectionHistory []string

	// Error state
	fatalErr string
}

// NewProjectForm creates a new project form that starts with stack selection.
// The stacks slice should exclude the "global" stack.
func NewProjectForm(stacks []registry.Stack, eng *engine.Engine, outputDir string) ProjectFormModel {
	// Build stack labels for the selection model
	stackLabels := make([]string, len(stacks))
	for i, s := range stacks {
		stackLabels[i] = s.Label
	}

	m := ProjectFormModel{
		stacks:          stacks,
		eng:             eng,
		outputDir:       outputDir,
		phase:           projectPhaseSelectStack,
		vars:            make(map[string]string),
		renderedContent: make(map[int]string),
		stackSelection:  NewSelection("Select a stack", stackLabels),
	}

	return m
}

// NewProjectFormWithVariant creates a new project form that skips selection
// and goes straight to variable collection for the given variant.
func NewProjectFormWithVariant(variant registry.Variant, eng *engine.Engine, outputDir string) ProjectFormModel {
	m := ProjectFormModel{
		variant:         variant,
		eng:             eng,
		outputDir:       outputDir,
		vars:            make(map[string]string),
		renderedContent: make(map[int]string),
	}

	if len(variant.Vars) > 0 {
		m.phase = projectPhaseCollectVars
		v := variant.Vars[0]
		m.textInput = NewTextInput(v.Prompt, v.Default, v.Required())
	} else {
		m.phase = projectPhaseProcessFiles
	}

	return m
}

// Init implements tea.Model.
func (m ProjectFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ProjectFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle ctrl+c and escape globally
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "escape":
			return m, tea.Quit
		}
	}

	switch m.phase {
	case projectPhaseSelectStack:
		return m.updateSelectStack(msg)
	case projectPhaseSelectVariant:
		return m.updateSelectVariant(msg)
	case projectPhaseCollectVars:
		return m.updateCollectVars(msg)
	case projectPhaseProcessFiles:
		return m.updateProcessFiles(msg)
	case projectPhaseConfirmOverwrite:
		return m.updateConfirmOverwrite(msg)
	case projectPhaseDone:
		return m, nil
	}

	return m, nil
}

func (m ProjectFormModel) updateSelectStack(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SelectionDoneMsg:
		m.selectedStack = msg.Index
		stack := m.stacks[msg.Index]

		// Save rendered view of completed selection
		m.selectionHistory = append(m.selectionHistory, m.stackSelection.viewString())

		if len(stack.Variants) == 1 {
			// Only one variant — skip variant selection
			m.variant = stack.Variants[0]
			return m.transitionToVarCollection()
		}

		// Multiple variants — show variant selection
		variantLabels := make([]string, len(stack.Variants))
		for i, v := range stack.Variants {
			variantLabels[i] = v.Label
		}
		m.variantSelection = NewSelection("Select a variant", variantLabels)
		m.phase = projectPhaseSelectVariant
		return m, nil

	default:
		var cmd tea.Cmd
		newSel, cmd := m.stackSelection.Update(msg)
		m.stackSelection = newSel.(SelectionModel)
		return m, cmd
	}
}

func (m ProjectFormModel) updateSelectVariant(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SelectionDoneMsg:
		m.selectedVariant = msg.Index
		stack := m.stacks[m.selectedStack]
		m.variant = stack.Variants[msg.Index]

		// Save rendered view of completed selection
		m.selectionHistory = append(m.selectionHistory, m.variantSelection.viewString())

		return m.transitionToVarCollection()

	default:
		var cmd tea.Cmd
		newSel, cmd := m.variantSelection.Update(msg)
		m.variantSelection = newSel.(SelectionModel)
		return m, cmd
	}
}

// transitionToVarCollection moves to variable collection or file processing.
func (m ProjectFormModel) transitionToVarCollection() (tea.Model, tea.Cmd) {
	if len(m.variant.Vars) > 0 {
		m.phase = projectPhaseCollectVars
		v := m.variant.Vars[0]
		m.textInput = NewTextInput(v.Prompt, v.Default, v.Required())
		return m, nil
	}

	m.phase = projectPhaseProcessFiles
	return m.processNextFile()
}

func (m ProjectFormModel) updateCollectVars(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TextInputDoneMsg:
		v := m.variant.Vars[m.varIndex]
		m.vars[v.Key] = msg.Value

		// Save rendered view of completed input
		m.varHistory = append(m.varHistory, m.textInput.viewString())

		// Advance to next variable or move to file processing
		m.varIndex++
		if m.varIndex < len(m.variant.Vars) {
			nextVar := m.variant.Vars[m.varIndex]
			m.textInput = NewTextInput(nextVar.Prompt, nextVar.Default, nextVar.Required())
		} else {
			m.phase = projectPhaseProcessFiles
			return m.processNextFile()
		}
		return m, nil

	default:
		var cmd tea.Cmd
		newInput, cmd := m.textInput.Update(msg)
		m.textInput = newInput.(TextInputModel)
		return m, cmd
	}
}

func (m ProjectFormModel) updateProcessFiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m ProjectFormModel) updateConfirmOverwrite(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ConfirmDoneMsg:
		fm := m.variant.Files[m.fileIndex]
		outputPath := filepath.Join(m.outputDir, fm.OutputPath)
		content := m.renderedContent[m.fileIndex]

		switch msg.Choice {
		case OverwriteYes:
			if err := m.eng.WriteFile(outputPath, content, true); err != nil {
				m.outcomes = append(m.outcomes, fileOutcome{
					outputPath: outputPath,
					errMsg:     err.Error(),
				})
			} else {
				m.outcomes = append(m.outcomes, fileOutcome{
					outputPath: outputPath,
					written:    true,
				})
			}

		case OverwriteNo:
			m.outcomes = append(m.outcomes, fileOutcome{
				outputPath: outputPath,
				skipped:    true,
			})

		case OverwriteAll:
			m.overwriteAll = true
			if err := m.eng.WriteFile(outputPath, content, true); err != nil {
				m.outcomes = append(m.outcomes, fileOutcome{
					outputPath: outputPath,
					errMsg:     err.Error(),
				})
			} else {
				m.outcomes = append(m.outcomes, fileOutcome{
					outputPath: outputPath,
					written:    true,
				})
			}
		}

		m.fileIndex++
		return m.processNextFile()

	default:
		var cmd tea.Cmd
		newConfirm, cmd := m.confirmModel.Update(msg)
		m.confirmModel = newConfirm.(ConfirmModel)
		return m, cmd
	}
}

func (m ProjectFormModel) processNextFile() (tea.Model, tea.Cmd) {
	for m.fileIndex < len(m.variant.Files) {
		fm := m.variant.Files[m.fileIndex]

		templatePath := filepath.Join(m.variant.StackKey, m.variant.VariantDir, fm.TemplatePath)
		content, err := m.eng.Render(templatePath, m.vars)
		if err != nil {
			m.fatalErr = fmt.Sprintf("Error rendering %s: %v", fm.TemplatePath, err)
			m.phase = projectPhaseDone
			return m, tea.Quit
		}

		m.renderedContent[m.fileIndex] = content

		outputPath := filepath.Join(m.outputDir, fm.OutputPath)

		_, statErr := os.Stat(outputPath)
		fileExists := statErr == nil

		if fileExists && !m.overwriteAll {
			m.phase = projectPhaseConfirmOverwrite
			m.confirmModel = NewConfirm(outputPath)
			return m, nil
		}

		if err := m.eng.WriteFile(outputPath, content, m.overwriteAll); err != nil {
			m.outcomes = append(m.outcomes, fileOutcome{
				outputPath: outputPath,
				errMsg:     err.Error(),
			})
		} else {
			m.outcomes = append(m.outcomes, fileOutcome{
				outputPath: outputPath,
				written:    true,
			})
		}

		m.fileIndex++
	}

	m.phase = projectPhaseDone
	return m, tea.Quit
}

// View implements tea.Model.
func (m ProjectFormModel) View() tea.View {
	return tea.NewView(m.viewString())
}

func (m ProjectFormModel) viewString() string {
	var b strings.Builder

	b.WriteString("\n")

	// Show completed selection history
	for _, h := range m.selectionHistory {
		b.WriteString(h)
	}

	// Show completed variable history
	for _, h := range m.varHistory {
		b.WriteString(h)
	}

	switch m.phase {
	case projectPhaseSelectStack:
		b.WriteString(m.stackSelection.viewString())

	case projectPhaseSelectVariant:
		b.WriteString(m.variantSelection.viewString())

	case projectPhaseCollectVars:
		b.WriteString(m.textInput.viewString())

	case projectPhaseConfirmOverwrite:
		b.WriteString("\n")
		b.WriteString(m.confirmModel.viewString())

	case projectPhaseDone:
		if m.fatalErr != "" {
			b.WriteString("\n")
			b.WriteString(ErrorStyle.Render(m.fatalErr))
			b.WriteString("\n")
		} else {
			b.WriteString("\n")
			for _, o := range m.outcomes {
				if o.errMsg != "" {
					b.WriteString(ErrorStyle.Render("  x "))
					b.WriteString(fmt.Sprintf("%s (%s)\n", o.outputPath, o.errMsg))
				} else if o.skipped {
					b.WriteString(SkipStyle.Render("  - "))
					b.WriteString(fmt.Sprintf("%s (skipped)\n", o.outputPath))
				} else if o.written {
					b.WriteString(SuccessStyle.Render("  + "))
					b.WriteString(fmt.Sprintf("%s\n", o.outputPath))
				}
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Outcomes returns the file outcomes (for testing).
func (m ProjectFormModel) Outcomes() []fileOutcome {
	return m.outcomes
}

// Variant returns the resolved variant (for testing).
func (m ProjectFormModel) Variant() registry.Variant {
	return m.variant
}

// Phase returns the current phase (for testing).
func (m ProjectFormModel) Phase() projectPhase {
	return m.phase
}
