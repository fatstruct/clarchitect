package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// TextInputDoneMsg is sent when the user submits their input.
type TextInputDoneMsg struct {
	Value string
}

// TextInputModel is a Bubbletea model for collecting a single text input.
type TextInputModel struct {
	prompt     string
	defaultVal string
	required   bool
	value      string
	cursorPos  int
	err        string
	done       bool
	submitted  string
}

// NewTextInput creates a new text input model.
func NewTextInput(prompt, defaultVal string, required bool) TextInputModel {
	return TextInputModel{
		prompt:     prompt,
		defaultVal: defaultVal,
		required:   required,
	}
}

// Init implements tea.Model.
func (m TextInputModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.value)
			if val == "" && m.defaultVal != "" {
				val = m.defaultVal
			}
			if val == "" && m.required {
				m.err = "This field is required"
				return m, nil
			}
			m.err = ""
			m.done = true
			m.submitted = val
			return m, func() tea.Msg { return TextInputDoneMsg{Value: val} }

		case "backspace":
			if m.cursorPos > 0 {
				runes := []rune(m.value)
				m.value = string(runes[:m.cursorPos-1]) + string(runes[m.cursorPos:])
				m.cursorPos--
			}
			m.err = ""

		case "left":
			if m.cursorPos > 0 {
				m.cursorPos--
			}

		case "right":
			if m.cursorPos < len([]rune(m.value)) {
				m.cursorPos++
			}

		case "home":
			m.cursorPos = 0

		case "end":
			m.cursorPos = len([]rune(m.value))

		case "ctrl+c":
			return m, tea.Quit

		default:
			// Insert typed rune at cursor position
			text := msg.String()
			if len(text) == 1 && text[0] >= 32 {
				runes := []rune(m.value)
				newRunes := make([]rune, 0, len(runes)+1)
				newRunes = append(newRunes, runes[:m.cursorPos]...)
				newRunes = append(newRunes, []rune(text)...)
				newRunes = append(newRunes, runes[m.cursorPos:]...)
				m.value = string(newRunes)
				m.cursorPos += len([]rune(text))
				m.err = ""
			}
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m TextInputModel) View() tea.View {
	return tea.NewView(m.viewString())
}

func (m TextInputModel) viewString() string {
	if m.done {
		return PromptStyle.Render(m.prompt+": ") + m.submitted + "\n"
	}

	var b strings.Builder

	b.WriteString(PromptStyle.Render(m.prompt + ": "))

	if m.defaultVal != "" {
		b.WriteString(DefaultHintStyle.Render("[" + m.defaultVal + "] "))
	}

	runes := []rune(m.value)
	before := string(runes[:m.cursorPos])
	after := ""
	cursor := " "
	if m.cursorPos < len(runes) {
		cursor = string(runes[m.cursorPos : m.cursorPos+1])
		after = string(runes[m.cursorPos+1:])
	}
	b.WriteString(before)
	b.WriteString(CursorStyle.Reverse(true).Render(cursor))
	b.WriteString(after)
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString(ErrorStyle.Render("  " + m.err))
		b.WriteString("\n")
	}

	return b.String()
}

// Done returns true if the input has been submitted.
func (m TextInputModel) Done() bool {
	return m.done
}

// Value returns the current text buffer.
func (m TextInputModel) Value() string {
	return m.value
}

// Submitted returns the final submitted value.
func (m TextInputModel) Submitted() string {
	return m.submitted
}
