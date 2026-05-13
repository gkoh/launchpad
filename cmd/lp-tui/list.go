package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gkoh/launchpad"
)

// bugTaskItem wraps a BugTask for the list.Model interface.
type bugTaskItem struct {
	task launchpad.BugTask
}

func (i bugTaskItem) Title() string {
	return i.task.Title
}

func (i bugTaskItem) Description() string {
	return fmt.Sprintf("%s  |  %s  |  %s",
		i.task.BugTargetDisplayName,
		i.task.Status,
		i.task.Importance,
	)
}

func (i bugTaskItem) FilterValue() string {
	return i.task.Title
}

// bugID extracts the bug ID from the BugLink URL.
func (i bugTaskItem) bugID() int {
	base := path.Base(i.task.BugLink)
	id, _ := strconv.Atoi(base)
	return id
}

type listModel struct {
	list   list.Model
	tasks  []launchpad.BugTask
	query  string // description of what was searched
	width  int
	height int
}

func newListModel(tasks []launchpad.BugTask, query string, width, height int) listModel {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = bugTaskItem{task: t}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		BorderLeftForeground(lipgloss.Color("205"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("245")).
		BorderLeftForeground(lipgloss.Color("205"))

	l := list.New(items, delegate, width, height-2)
	l.Title = query
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	return listModel{
		list:   l,
		tasks:  tasks,
		query:  query,
		width:  width,
		height: height,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[Enter] view  [/] search  [Esc] back  [Ctrl+C] quit"))
	return b.String()
}

// selectedBugID returns the bug ID of the currently selected item, or 0.
func (m listModel) selectedBugID() int {
	item, ok := m.list.SelectedItem().(bugTaskItem)
	if !ok {
		return 0
	}
	return item.bugID()
}
