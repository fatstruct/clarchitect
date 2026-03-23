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

// formPhase tracks the current phase of the global form.
type formPhase int

const (
	phaseCollectVars formPhase = iota
	phaseProcessFiles
	phaseConfirmOverwrite
	phaseDone
)

// fileOutcome records what happened to each file.
type fileOutcome struct {
	outputPath string
	written    bool
	skipped    bool
	errMsg     string
}

// GlobalFormModel orchestrates variable collection, file rendering, and overwrite confirmation.
type GlobalFormModel struct {
	variant   registry.Variant
	eng       *engine.Engine
	outputDir string

	// Variable collection
	phase      formPhase
	varIndex   int                 // Index into variant.Vars for current input
	textInput  TextInputModel      // Current text input model
	vars       map[string]string   // Collected variable values
	varHistory []string            // Rendered views of completed inputs

	// File processing
	fileIndex    int             // Index into variant.Files for current file
	overwriteAll bool           // If true, skip confirmation for remaining files
	confirmModel ConfirmModel   // Current confirmation model
	outcomes     []fileOutcome  // Results for each file

	// Rendered content cache (file index -> rendered content)
	renderedContent map[int]string

	// Error state
	fatalErr string
}

// NewGlobalForm creates a new global form model.
func NewGlobalForm(variant registry.Variant, eng *engine.Engine, outputDir string) GlobalFormModel {
	m := GlobalFormModel{
		variant:         variant,
		eng:             eng,
		outputDir:       outputDir,
		vars:            make(map[string]string),
		renderedContent: make(map[int]string),
	}

	if len(variant.Vars) > 0 {
		m.phase = phaseCollectVars
		v := variant.Vars[0]
		m.textInput = NewTextInput(v.Prompt, v.Default, v.Required())
	} else {
		m.phase = phaseProcessFiles
	}

	return m
}

// Init implements tea.Model.
func (m GlobalFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m GlobalFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle ctrl+c and escape in any phase
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "escape":
			return m, tea.Quit
		}
	}

	switch m.phase {
	case phaseCollectVars:
		return m.updateCollectVars(msg)
	case phaseProcessFiles:
		return m.updateProcessFiles(msg)
	case phaseConfirmOverwrite:
		return m.updateConfirmOverwrite(msg)
	case phaseDone:
		return m, nil
	}

	return m, nil
}

func (m GlobalFormModel) updateCollectVars(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TextInputDoneMsg:
		// Store the collected value
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
			m.phase = phaseProcessFiles
			return m.processNextFile()
		}
		return m, nil

	default:
		// Forward to text input
		var cmd tea.Cmd
		newInput, cmd := m.textInput.Update(msg)
		m.textInput = newInput.(TextInputModel)
		return m, cmd
	}
}

func (m GlobalFormModel) updateProcessFiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	// This phase is entered via processNextFile, which handles file logic directly.
	// We shouldn't normally receive messages here unless something unexpected happens.
	return m, nil
}

func (m GlobalFormModel) updateConfirmOverwrite(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Forward to confirm model
		var cmd tea.Cmd
		newConfirm, cmd := m.confirmModel.Update(msg)
		m.confirmModel = newConfirm.(ConfirmModel)
		return m, cmd
	}
}

// processNextFile processes the next file in the variant's file list.
// It renders the template, checks if the file exists, and either writes it
// or moves to the confirmation phase.
func (m GlobalFormModel) processNextFile() (tea.Model, tea.Cmd) {
	for m.fileIndex < len(m.variant.Files) {
		fm := m.variant.Files[m.fileIndex]

		// Build the full template path
		templatePath := filepath.Join(m.variant.StackKey, m.variant.VariantDir, fm.TemplatePath)
		content, err := m.eng.Render(templatePath, m.vars)
		if err != nil {
			m.fatalErr = fmt.Sprintf("Error rendering %s: %v", fm.TemplatePath, err)
			m.phase = phaseDone
			return m, tea.Quit
		}

		// Cache rendered content
		m.renderedContent[m.fileIndex] = content

		outputPath := filepath.Join(m.outputDir, fm.OutputPath)

		// Check if file exists
		_, statErr := os.Stat(outputPath)
		fileExists := statErr == nil

		if fileExists && !m.overwriteAll {
			// Need confirmation
			m.phase = phaseConfirmOverwrite
			m.confirmModel = NewConfirm(outputPath)
			return m, nil
		}

		// Write the file (force=true if overwriteAll, otherwise file doesn't exist)
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

	// All files processed
	m.phase = phaseDone
	return m, tea.Quit
}

// View implements tea.Model.
func (m GlobalFormModel) View() tea.View {
	return tea.NewView(m.viewString())
}

func (m GlobalFormModel) viewString() string {
	var b strings.Builder

	b.WriteString("\n")

	// Show completed input history
	for _, h := range m.varHistory {
		b.WriteString(h)
	}

	switch m.phase {
	case phaseCollectVars:
		b.WriteString(m.textInput.viewString())

	case phaseConfirmOverwrite:
		b.WriteString("\n")
		b.WriteString(m.confirmModel.viewString())

	case phaseDone:
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
func (m GlobalFormModel) Outcomes() []fileOutcome {
	return m.outcomes
}
