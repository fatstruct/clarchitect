package tui

import "charm.land/lipgloss/v2"

// Styles for the TUI. lipgloss v2 uses color.Color directly.
// NO_COLOR is respected automatically by the lipgloss Writer.
var (
	PromptStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7B78EE"))
	DefaultHintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	ErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666"))
	SuccessStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#66FF66"))
	SkipStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC66"))
	CursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7B78EE"))
)
