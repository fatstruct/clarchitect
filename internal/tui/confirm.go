package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// OverwriteChoice represents the user's overwrite decision.
type OverwriteChoice int

const (
	OverwriteYes OverwriteChoice = iota
	OverwriteNo
	OverwriteAll
)

// ConfirmDoneMsg is sent when the user makes an overwrite decision.
type ConfirmDoneMsg struct {
	Choice OverwriteChoice
}

// ConfirmModel is a Bubbletea model for overwrite confirmation.
type ConfirmModel struct {
	filePath string
	done     bool
	choice   OverwriteChoice
}

// NewConfirm creates a new overwrite confirmation model.
func NewConfirm(filePath string) ConfirmModel {
	return ConfirmModel{filePath: filePath}
}

// Init implements tea.Model.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch strings.ToLower(msg.String()) {
		case "y":
			m.done = true
			m.choice = OverwriteYes
			return m, func() tea.Msg { return ConfirmDoneMsg{Choice: OverwriteYes} }
		case "n":
			m.done = true
			m.choice = OverwriteNo
			return m, func() tea.Msg { return ConfirmDoneMsg{Choice: OverwriteNo} }
		case "a":
			m.done = true
			m.choice = OverwriteAll
			return m, func() tea.Msg { return ConfirmDoneMsg{Choice: OverwriteAll} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m ConfirmModel) View() tea.View {
	return tea.NewView(m.viewString())
}

func (m ConfirmModel) viewString() string {
	if m.done {
		var label string
		switch m.choice {
		case OverwriteYes:
			label = "yes"
		case OverwriteNo:
			label = "no"
		case OverwriteAll:
			label = "all"
		}
		return PromptStyle.Render("Overwrite? ") + label + "\n"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("File already exists: %s\n", SkipStyle.Render(m.filePath)))
	b.WriteString(PromptStyle.Render("Overwrite? "))
	b.WriteString(DefaultHintStyle.Render("[y]es / [n]o / [a]ll"))
	b.WriteString("\n")
	return b.String()
}

// Done returns true if the user has made a choice.
func (m ConfirmModel) Done() bool {
	return m.done
}

// Choice returns the selected overwrite choice.
func (m ConfirmModel) Choice() OverwriteChoice {
	return m.choice
}
