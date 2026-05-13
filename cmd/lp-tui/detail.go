package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gkoh/launchpad"
)

type detailModel struct {
	viewport viewport.Model
	bug      *launchpad.Bug
	tasks    []launchpad.BugTask
	assignees map[string]string
	messages []launchpad.Message
	owners   map[string]string
	ready    bool
	width    int
	height   int
}

func newDetailModel(width, height int) detailModel {
	return detailModel{
		width:  width,
		height: height,
	}
}

func (m detailModel) Init() tea.Cmd {
	return nil
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.ready {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
		}
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m detailModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		fmt.Sprintf("[↑/↓/PgUp/PgDn] scroll  [Esc] back  [Ctrl+C] quit  (%d%%)",
			int(m.viewport.ScrollPercent()*100)),
	))
	return b.String()
}

// setBug sets the bug and rebuilds the viewport content.
func (m *detailModel) setBug(bug *launchpad.Bug) {
	m.bug = bug
	m.rebuild()
}

// setTasks sets the tasks and assignees and rebuilds.
func (m *detailModel) setTasks(tasks []launchpad.BugTask, assignees map[string]string) {
	m.tasks = tasks
	m.assignees = assignees
	m.rebuild()
}

// setComments sets the messages and owners and rebuilds.
func (m *detailModel) setComments(messages []launchpad.Message, owners map[string]string) {
	m.messages = messages
	m.owners = owners
	m.rebuild()
}

// rebuild regenerates the viewport content from current data.
func (m *detailModel) rebuild() {
	if m.bug == nil {
		return
	}

	content := m.renderContent()

	if !m.ready {
		m.viewport = viewport.New(m.width, m.height-2)
		m.viewport.SetContent(content)
		m.ready = true
	} else {
		m.viewport.SetContent(content)
	}
}

// reset clears all detail data.
func (m *detailModel) reset() {
	m.bug = nil
	m.tasks = nil
	m.assignees = nil
	m.messages = nil
	m.owners = nil
	m.ready = false
}

const (
	maxDescriptionLength = 500
	maxCommentLength     = 500
)

func (m detailModel) renderContent() string {
	bug := m.bug
	var b strings.Builder

	// Title.
	b.WriteString(titleStyle.Render(fmt.Sprintf("Bug #%d: %s", bug.ID, bug.Title)))
	b.WriteString("\n\n")

	// Fields.
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Info type:"), bug.InformationType))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Lock status:"), bug.LockStatus))
	b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Heat:"), bug.Heat))
	if len(bug.Tags) > 0 {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Tags:"), strings.Join(bug.Tags, ", ")))
	}
	if bug.DateCreated != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Created:"), bug.DateCreated.Format("2006-01-02 15:04:05")))
	}
	if bug.DateLastUpdated != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Updated:"), bug.DateLastUpdated.Format("2006-01-02 15:04:05")))
	}
	b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Messages:"), bug.MessageCount))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Web:"), bug.WebLink))

	// Description.
	if bug.Description != "" {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Description"))
		b.WriteString("\n\n")
		desc := bug.Description
		if len(desc) > maxDescriptionLength {
			desc = desc[:maxDescriptionLength] + "..."
		}
		b.WriteString(desc)
		b.WriteString("\n")
	}

	// Tasks.
	if len(m.tasks) > 0 {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Tasks"))
		b.WriteString("\n")
		for _, task := range m.tasks {
			b.WriteString(fmt.Sprintf("\n  %s\n", labelStyle.Render(string(task.BugTargetDisplayName))))
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Status:"), task.Status))
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Importance:"), task.Importance))
			assignee := "unassigned"
			if name, ok := m.assignees[task.AssigneeLink.String()]; ok && !task.AssigneeLink.IsZero() {
				assignee = name
			}
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Assignee:"), assignee))
		}
	}

	// Comments.
	if len(m.messages) > 0 {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("Comments (%d)", len(m.messages))))
		b.WriteString("\n")
		for i, msg := range m.messages {
			owner := msg.OwnerLink.String()
			if name, ok := m.owners[msg.OwnerLink.String()]; ok {
				owner = name
			}
			date := "unknown"
			if msg.DateCreated != nil {
				date = msg.DateCreated.Format("2006-01-02 15:04:05")
			}
			b.WriteString(fmt.Sprintf("\n  %s\n", labelStyle.Render(fmt.Sprintf("#%d by %s on %s", i+1, owner, date))))
			content := msg.Content
			if len(content) > maxCommentLength {
				content = content[:maxCommentLength] + "..."
			}
			if content != "" {
				for _, line := range strings.Split(content, "\n") {
					b.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}
		}
	}

	return b.String()
}
