package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// sendKey sends a special key press to the model.
func sendKey(m tea.Model, code rune) (tea.Model, tea.Cmd) {
	msg := tea.KeyPressMsg{Code: code}
	return m.Update(msg)
}

// sendRune sends a printable character key press to the model.
func sendRune(m tea.Model, r rune) (tea.Model, tea.Cmd) {
	msg := tea.KeyPressMsg{Code: r, Text: string(r)}
	return m.Update(msg)
}

// typeString sends a sequence of rune key presses to the model.
func typeString(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = sendRune(m, r)
	}
	return m
}

func TestTextInputAcceptsTypedCharacters(t *testing.T) {
	m := NewTextInput("Name", "", true)
	var model tea.Model = m

	model = typeString(model, "Alice")

	ti := model.(TextInputModel)
	if ti.Value() != "Alice" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "Alice")
	}
	if ti.Done() {
		t.Error("expected Done() to be false before Enter")
	}
}

func TestTextInputEnterOnEmptyRequiredShowsError(t *testing.T) {
	m := NewTextInput("Name", "", true)
	var model tea.Model = m

	model, _ = sendKey(model, tea.KeyEnter)

	ti := model.(TextInputModel)
	if ti.Done() {
		t.Error("expected Done() to be false for empty required field")
	}

	view := ti.viewString()
	if !strings.Contains(view, "required") {
		t.Errorf("expected error message containing 'required', got view:\n%s", view)
	}
}

func TestTextInputEnterOnEmptyOptionalReturnsDefault(t *testing.T) {
	m := NewTextInput("Style", "test alongside implementation", false)
	var model tea.Model = m

	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)

	ti := model.(TextInputModel)
	if !ti.Done() {
		t.Error("expected Done() to be true after Enter on optional field")
	}
	if ti.Submitted() != "test alongside implementation" {
		t.Errorf("Submitted() = %q, want %q", ti.Submitted(), "test alongside implementation")
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(TextInputDoneMsg)
	if !ok {
		t.Fatalf("expected TextInputDoneMsg, got %T", msg)
	}
	if doneMsg.Value != "test alongside implementation" {
		t.Errorf("TextInputDoneMsg.Value = %q, want %q", doneMsg.Value, "test alongside implementation")
	}
}

func TestTextInputEnterWithTextReturnsText(t *testing.T) {
	m := NewTextInput("Name", "", true)
	var model tea.Model = m

	model = typeString(model, "Bob")
	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)

	ti := model.(TextInputModel)
	if !ti.Done() {
		t.Error("expected Done() to be true after Enter with text")
	}
	if ti.Submitted() != "Bob" {
		t.Errorf("Submitted() = %q, want %q", ti.Submitted(), "Bob")
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(TextInputDoneMsg)
	if !ok {
		t.Fatalf("expected TextInputDoneMsg, got %T", msg)
	}
	if doneMsg.Value != "Bob" {
		t.Errorf("TextInputDoneMsg.Value = %q, want %q", doneMsg.Value, "Bob")
	}
}

func TestTextInputEnterWithTextOverridesDefault(t *testing.T) {
	m := NewTextInput("Style", "default-style", false)
	var model tea.Model = m

	model = typeString(model, "custom-style")
	model, _ = sendKey(model, tea.KeyEnter)

	ti := model.(TextInputModel)
	if !ti.Done() {
		t.Error("expected Done() to be true")
	}
	if ti.Submitted() != "custom-style" {
		t.Errorf("Submitted() = %q, want %q", ti.Submitted(), "custom-style")
	}
}

func TestTextInputBackspaceDeletesCharacter(t *testing.T) {
	m := NewTextInput("Name", "", true)
	var model tea.Model = m

	model = typeString(model, "Test")
	model, _ = sendKey(model, tea.KeyBackspace)

	ti := model.(TextInputModel)
	if ti.Value() != "Tes" {
		t.Errorf("Value() after backspace = %q, want %q", ti.Value(), "Tes")
	}
}

func TestTextInputViewShowsDefaultHint(t *testing.T) {
	m := NewTextInput("Style", "my-default", false)
	view := m.viewString()

	if !strings.Contains(view, "[my-default]") {
		t.Errorf("expected view to contain default hint [my-default], got:\n%s", view)
	}
}

func TestTextInputViewShowsPrompt(t *testing.T) {
	m := NewTextInput("Your name", "", true)
	view := m.viewString()

	if !strings.Contains(view, "Your name") {
		t.Errorf("expected view to contain prompt 'Your name', got:\n%s", view)
	}
}
