package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SelectionDoneMsg is sent when the user confirms a selection.
type SelectionDoneMsg struct {
	Index int
}

// SelectionModel is a Bubbletea model for selecting from a numbered list.
type SelectionModel struct {
	title   string
	items   []string
	cursor  int
	done    bool
	chosen  int
}

// Styles for the selection list.
var (
	SelectionCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7B78EE")).Bold(true)
	SelectionItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	SelectionTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7B78EE"))
)

// NewSelection creates a new selection model with the given title and items.
func NewSelection(title string, items []string) SelectionModel {
	return SelectionModel{
		title: title,
		items: items,
	}
}

// Init implements tea.Model.
func (m SelectionModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.done = true
			m.chosen = m.cursor
			return m, func() tea.Msg { return SelectionDoneMsg{Index: m.cursor} }
		case "ctrl+c", "escape":
			return m, tea.Quit
		default:
			// Number keys 1-9 for quick selection
			if len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9' {
				idx := int(msg.String()[0] - '1')
				if idx < len(m.items) {
					m.cursor = idx
					m.done = true
					m.chosen = idx
					return m, func() tea.Msg { return SelectionDoneMsg{Index: idx} }
				}
			}
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m SelectionModel) View() tea.View {
	return tea.NewView(m.viewString())
}

func (m SelectionModel) viewString() string {
	if m.done {
		return SelectionTitleStyle.Render(m.title+": ") + m.items[m.chosen] + "\n"
	}

	var b strings.Builder

	b.WriteString(SelectionTitleStyle.Render(m.title))
	b.WriteString("\n")

	for i, item := range m.items {
		num := fmt.Sprintf("%d. ", i+1)
		if i == m.cursor {
			b.WriteString(SelectionCursorStyle.Render("> "+num+item))
		} else {
			b.WriteString(SelectionItemStyle.Render("  "+num+item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(DefaultHintStyle.Render("  Use arrows/j/k to navigate, 1-9 or Enter to select"))
	b.WriteString("\n")

	return b.String()
}

// Done returns true if a selection has been made.
func (m SelectionModel) Done() bool {
	return m.done
}

// Chosen returns the index of the selected item.
func (m SelectionModel) Chosen() int {
	return m.chosen
}
