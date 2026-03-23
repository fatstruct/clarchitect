package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestConfirmYes(t *testing.T) {
	m := NewConfirm("/home/user/.claude/CLAUDE.md")
	var model tea.Model = m

	var cmd tea.Cmd
	model, cmd = sendRune(model, 'y')

	cm := model.(ConfirmModel)
	if !cm.Done() {
		t.Error("expected Done() to be true after pressing 'y'")
	}
	if cm.Choice() != OverwriteYes {
		t.Errorf("Choice() = %d, want OverwriteYes (%d)", cm.Choice(), OverwriteYes)
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(ConfirmDoneMsg)
	if !ok {
		t.Fatalf("expected ConfirmDoneMsg, got %T", msg)
	}
	if doneMsg.Choice != OverwriteYes {
		t.Errorf("ConfirmDoneMsg.Choice = %d, want OverwriteYes", doneMsg.Choice)
	}
}

func TestConfirmNo(t *testing.T) {
	m := NewConfirm("/home/user/.claude/CLAUDE.md")
	var model tea.Model = m

	var cmd tea.Cmd
	model, cmd = sendRune(model, 'n')

	cm := model.(ConfirmModel)
	if !cm.Done() {
		t.Error("expected Done() to be true after pressing 'n'")
	}
	if cm.Choice() != OverwriteNo {
		t.Errorf("Choice() = %d, want OverwriteNo (%d)", cm.Choice(), OverwriteNo)
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(ConfirmDoneMsg)
	if !ok {
		t.Fatalf("expected ConfirmDoneMsg, got %T", msg)
	}
	if doneMsg.Choice != OverwriteNo {
		t.Errorf("ConfirmDoneMsg.Choice = %d, want OverwriteNo", doneMsg.Choice)
	}
}

func TestConfirmAll(t *testing.T) {
	m := NewConfirm("/home/user/.claude/CLAUDE.md")
	var model tea.Model = m

	var cmd tea.Cmd
	model, cmd = sendRune(model, 'a')

	cm := model.(ConfirmModel)
	if !cm.Done() {
		t.Error("expected Done() to be true after pressing 'a'")
	}
	if cm.Choice() != OverwriteAll {
		t.Errorf("Choice() = %d, want OverwriteAll (%d)", cm.Choice(), OverwriteAll)
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(ConfirmDoneMsg)
	if !ok {
		t.Fatalf("expected ConfirmDoneMsg, got %T", msg)
	}
	if doneMsg.Choice != OverwriteAll {
		t.Errorf("ConfirmDoneMsg.Choice = %d, want OverwriteAll", doneMsg.Choice)
	}
}

func TestConfirmIgnoresUnrecognizedKeys(t *testing.T) {
	m := NewConfirm("/some/path")
	var model tea.Model = m

	model, _ = sendRune(model, 'x')

	cm := model.(ConfirmModel)
	if cm.Done() {
		t.Error("expected Done() to be false for unrecognized key")
	}
}

func TestConfirmUppercaseKeysWork(t *testing.T) {
	m := NewConfirm("/some/path")
	var model tea.Model = m

	model, _ = sendRune(model, 'Y')

	cm := model.(ConfirmModel)
	if !cm.Done() {
		t.Error("expected Done() to be true for uppercase 'Y'")
	}
	if cm.Choice() != OverwriteYes {
		t.Errorf("Choice() = %d, want OverwriteYes", cm.Choice())
	}
}

func TestConfirmViewShowsFilePath(t *testing.T) {
	m := NewConfirm("/home/user/.claude/CLAUDE.md")
	view := m.viewString()

	if !strings.Contains(view, "/home/user/.claude/CLAUDE.md") {
		t.Errorf("expected view to contain file path, got:\n%s", view)
	}
	if !strings.Contains(view, "[y]es") {
		t.Errorf("expected view to contain '[y]es', got:\n%s", view)
	}
}

func TestConfirmDoneViewShowsChoice(t *testing.T) {
	m := NewConfirm("/some/path")
	var model tea.Model = m

	model, _ = sendRune(model, 'n')

	cm := model.(ConfirmModel)
	view := cm.viewString()

	if !strings.Contains(view, "no") {
		t.Errorf("expected done view to contain 'no', got:\n%s", view)
	}
}
