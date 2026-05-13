package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	inputProject = iota
	inputSearch
	numInputs
)

type searchModel struct {
	inputs  [numInputs]textinput.Model
	focused int
}

func newSearchModel() searchModel {
	var inputs [numInputs]textinput.Model

	inputs[inputProject] = textinput.New()
	inputs[inputProject].Placeholder = "e.g. ubuntu, linux"
	inputs[inputProject].Prompt = focusedInputStyle.Render("Project: ")
	inputs[inputProject].Focus()
	inputs[inputProject].CharLimit = 100

	inputs[inputSearch] = textinput.New()
	inputs[inputSearch].Placeholder = "search text or bug ID"
	inputs[inputSearch].Prompt = blurredInputStyle.Render("Search:  ")
	inputs[inputSearch].CharLimit = 200

	return searchModel{
		inputs:  inputs,
		focused: inputProject,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			m.focused = (m.focused + 1) % numInputs
			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Prompt = focusedInputStyle.Render(m.promptLabel(i))
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Prompt = blurredInputStyle.Render(m.promptLabel(i))
					m.inputs[i].Blur()
				}
			}
			return m, nil
		}
	}

	// Update the focused input.
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m searchModel) promptLabel(i int) string {
	switch i {
	case inputProject:
		return "Project: "
	case inputSearch:
		return "Search:  "
	}
	return ""
}

func (m searchModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Launchpad Bug Browser"))
	b.WriteString("\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[Tab] switch field  [Enter] search  [Ctrl+C] quit"))

	return b.String()
}

// project returns the trimmed project name.
func (m searchModel) project() string {
	return strings.TrimSpace(m.inputs[inputProject].Value())
}

// searchText returns the trimmed search text.
func (m searchModel) searchText() string {
	return strings.TrimSpace(m.inputs[inputSearch].Value())
}

// isDirectBugID returns true and the bug ID if the search text is a pure integer.
func (m searchModel) isDirectBugID() (int, bool) {
	text := m.searchText()
	if text == "" {
		return 0, false
	}
	// Strip leading # for convenience (e.g. "#12345").
	text = strings.TrimPrefix(text, "#")
	id, err := strconv.Atoi(text)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// validate returns an error message if the inputs are invalid.
func (m searchModel) validate() string {
	if bugID, ok := m.isDirectBugID(); ok && bugID > 0 {
		return "" // direct bug ID doesn't need project
	}
	if m.project() == "" {
		return "Project is required for search"
	}
	return ""
}

// reset clears search text but keeps project.
func (m *searchModel) reset() {
	m.inputs[inputSearch].SetValue("")
}

// statusLine returns a summary of what will be searched.
func (m searchModel) statusLine() string {
	if bugID, ok := m.isDirectBugID(); ok {
		return fmt.Sprintf("Fetching bug #%d...", bugID)
	}
	text := m.searchText()
	if text != "" {
		return fmt.Sprintf("Searching %s for %q...", m.project(), text)
	}
	return fmt.Sprintf("Searching %s...", m.project())
}
