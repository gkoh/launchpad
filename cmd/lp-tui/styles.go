package main

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle is used for main headings.
	titleStyle = lipgloss.NewStyle().Bold(true)

	// subtitleStyle is used for section headings.
	subtitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))

	// labelStyle is used for field labels in detail view.
	labelStyle = lipgloss.NewStyle().Bold(true)

	// helpStyle is used for the help bar at the bottom.
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// errorStyle is used for error messages.
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// spinnerStyle is used for the loading spinner.
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// focusedInputStyle is used for the focused text input prompt.
	focusedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// blurredInputStyle is used for unfocused text input prompts.
	blurredInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
