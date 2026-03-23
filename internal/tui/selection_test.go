package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSelectionInitialState(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)

	if m.Done() {
		t.Error("expected Done() to be false initially")
	}
	if m.Chosen() != 0 {
		t.Errorf("Chosen() = %d, want 0", m.Chosen())
	}

	view := m.viewString()
	if !strings.Contains(view, "Pick one") {
		t.Errorf("expected view to contain title, got:\n%s", view)
	}
	for _, item := range items {
		if !strings.Contains(view, item) {
			t.Errorf("expected view to contain %q, got:\n%s", item, view)
		}
	}
}

func TestSelectionDownArrowMovesCursor(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	model, _ = sendKey(model, tea.KeyDown)

	sm := model.(SelectionModel)
	if sm.cursor != 1 {
		t.Errorf("cursor = %d after down, want 1", sm.cursor)
	}

	model, _ = sendKey(model, tea.KeyDown)
	sm = model.(SelectionModel)
	if sm.cursor != 2 {
		t.Errorf("cursor = %d after second down, want 2", sm.cursor)
	}

	// Should not go beyond last item
	model, _ = sendKey(model, tea.KeyDown)
	sm = model.(SelectionModel)
	if sm.cursor != 2 {
		t.Errorf("cursor = %d after down at bottom, want 2", sm.cursor)
	}
}

func TestSelectionUpArrowMovesCursor(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	// Move down first, then up
	model, _ = sendKey(model, tea.KeyDown)
	model, _ = sendKey(model, tea.KeyDown)
	model, _ = sendKey(model, tea.KeyUp)

	sm := model.(SelectionModel)
	if sm.cursor != 1 {
		t.Errorf("cursor = %d after up, want 1", sm.cursor)
	}

	// Should not go above first item
	model, _ = sendKey(model, tea.KeyUp)
	model, _ = sendKey(model, tea.KeyUp)
	sm = model.(SelectionModel)
	if sm.cursor != 0 {
		t.Errorf("cursor = %d after up at top, want 0", sm.cursor)
	}
}

func TestSelectionJKNavigation(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	// j moves down
	model, _ = sendRune(model, 'j')
	sm := model.(SelectionModel)
	if sm.cursor != 1 {
		t.Errorf("cursor = %d after j, want 1", sm.cursor)
	}

	// k moves up
	model, _ = sendRune(model, 'k')
	sm = model.(SelectionModel)
	if sm.cursor != 0 {
		t.Errorf("cursor = %d after k, want 0", sm.cursor)
	}
}

func TestSelectionEnterConfirms(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	// Move to Beta, then confirm
	model, _ = sendKey(model, tea.KeyDown)
	var cmd tea.Cmd
	model, cmd = sendKey(model, tea.KeyEnter)

	sm := model.(SelectionModel)
	if !sm.Done() {
		t.Error("expected Done() to be true after Enter")
	}
	if sm.Chosen() != 1 {
		t.Errorf("Chosen() = %d, want 1", sm.Chosen())
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(SelectionDoneMsg)
	if !ok {
		t.Fatalf("expected SelectionDoneMsg, got %T", msg)
	}
	if doneMsg.Index != 1 {
		t.Errorf("SelectionDoneMsg.Index = %d, want 1", doneMsg.Index)
	}
}

func TestSelectionNumberKeySelectsDirectly(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	var cmd tea.Cmd
	model, cmd = sendRune(model, '2')

	sm := model.(SelectionModel)
	if !sm.Done() {
		t.Error("expected Done() to be true after number key")
	}
	if sm.Chosen() != 1 {
		t.Errorf("Chosen() = %d, want 1 (for key '2')", sm.Chosen())
	}

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	msg := cmd()
	doneMsg, ok := msg.(SelectionDoneMsg)
	if !ok {
		t.Fatalf("expected SelectionDoneMsg, got %T", msg)
	}
	if doneMsg.Index != 1 {
		t.Errorf("SelectionDoneMsg.Index = %d, want 1", doneMsg.Index)
	}
}

func TestSelectionNumberKeyOutOfRangeIgnored(t *testing.T) {
	items := []string{"Alpha", "Beta"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	model, _ = sendRune(model, '5')

	sm := model.(SelectionModel)
	if sm.Done() {
		t.Error("expected Done() to be false for out-of-range number key")
	}
}

func TestSelectionDoneViewShowsChoice(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	model, _ = sendRune(model, '1')

	sm := model.(SelectionModel)
	view := sm.viewString()
	if !strings.Contains(view, "Alpha") {
		t.Errorf("expected done view to contain chosen item, got:\n%s", view)
	}
	if !strings.Contains(view, "Pick one") {
		t.Errorf("expected done view to contain title, got:\n%s", view)
	}
}

func TestSelectionIgnoresInputAfterDone(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	m := NewSelection("Pick one", items)
	var model tea.Model = m

	model, _ = sendRune(model, '1')
	model, _ = sendRune(model, '3')

	sm := model.(SelectionModel)
	if sm.Chosen() != 0 {
		t.Errorf("Chosen() = %d, want 0 (should not change after done)", sm.Chosen())
	}
}
