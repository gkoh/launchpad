package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gkoh/launchpad"
)

// viewState tracks which view is active.
type viewState int

const (
	viewSearch viewState = iota
	viewList
	viewDetail
)

// model is the top-level Bubble Tea model.
type model struct {
	client  *launchpad.Client
	view    viewState
	prevView viewState // for back navigation from detail
	search  searchModel
	list    listModel
	detail  detailModel
	spinner spinner.Model
	loading bool
	status  string // status/error message
	width   int
	height  int
}

func newModel(client *launchpad.Client) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return model{
		client:  client,
		view:    viewSearch,
		search:  newSearchModel(),
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.search.Init(),
		m.spinner.Tick,
	)
}

// --- Messages ---

type searchResultsMsg struct {
	tasks []launchpad.BugTask
	query string
	err   error
}

type bugDetailMsg struct {
	bug *launchpad.Bug
	err error
}

type bugTasksMsg struct {
	tasks     []launchpad.BugTask
	assignees map[string]string
	err       error
}

type commentsMsg struct {
	messages []launchpad.Message
	owners   map[string]string
	err      error
}

// --- Commands ---

func searchBugsCmd(client *launchpad.Client, project, text string) tea.Cmd {
	return func() tea.Msg {
		params := url.Values{}
		params.Set("ws.op", "searchTasks")
		if text != "" {
			params.Set("search_text", text)
		}
		params.Set("ws.size", "50")

		path := fmt.Sprintf("/%s?%s", project, params.Encode())
		resp, err := client.Get(path)
		if err != nil {
			return searchResultsMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return searchResultsMsg{err: err}
		}

		if resp.StatusCode == http.StatusNotFound {
			return searchResultsMsg{err: fmt.Errorf("project %q not found", project)}
		}
		if resp.StatusCode != http.StatusOK {
			return searchResultsMsg{err: fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
		}

		var collection launchpad.BugTaskCollection
		if err := json.Unmarshal(body, &collection); err != nil {
			return searchResultsMsg{err: err}
		}

		query := fmt.Sprintf("Results for %s", project)
		if text != "" {
			query = fmt.Sprintf("Results for %s: %q", project, text)
		}

		return searchResultsMsg{tasks: collection.Entries, query: query}
	}
}

func fetchBugCmd(client *launchpad.Client, bugID int) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.Get(fmt.Sprintf("/bugs/%d", bugID))
		if err != nil {
			return bugDetailMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return bugDetailMsg{err: err}
		}

		if resp.StatusCode == http.StatusNotFound {
			return bugDetailMsg{err: fmt.Errorf("bug #%d not found", bugID)}
		}
		if resp.StatusCode != http.StatusOK {
			return bugDetailMsg{err: fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
		}

		var bug launchpad.Bug
		if err := json.Unmarshal(body, &bug); err != nil {
			return bugDetailMsg{err: err}
		}

		return bugDetailMsg{bug: &bug}
	}
}

func fetchBugTasksCmd(client *launchpad.Client, tasksURL string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest(http.MethodGet, tasksURL, nil)
		if err != nil {
			return bugTasksMsg{err: err}
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return bugTasksMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return bugTasksMsg{err: err}
		}

		if resp.StatusCode != http.StatusOK {
			return bugTasksMsg{err: fmt.Errorf("bug tasks returned %d", resp.StatusCode)}
		}

		var collection launchpad.BugTaskCollection
		if err := json.Unmarshal(body, &collection); err != nil {
			return bugTasksMsg{err: err}
		}

		// Resolve assignees.
		assignees := resolvePersonLinks(client, uniqueLinks(collection.Entries, func(t launchpad.BugTask) string {
			return t.AssigneeLink
		}))

		return bugTasksMsg{tasks: collection.Entries, assignees: assignees}
	}
}

func fetchCommentsCmd(client *launchpad.Client, messagesURL string) tea.Cmd {
	return func() tea.Msg {
		var allMessages []launchpad.Message
		url := messagesURL

		for url != "" {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return commentsMsg{err: err}
			}
			req.Header.Set("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return commentsMsg{err: err}
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return commentsMsg{err: err}
			}

			if resp.StatusCode != http.StatusOK {
				return commentsMsg{err: fmt.Errorf("messages returned %d", resp.StatusCode)}
			}

			var collection launchpad.MessageCollection
			if err := json.Unmarshal(body, &collection); err != nil {
				return commentsMsg{err: err}
			}

			allMessages = append(allMessages, collection.Entries...)
			url = collection.NextCollectionLink
		}

		// Resolve owners.
		owners := resolvePersonLinks(client, uniqueLinks(allMessages, func(m launchpad.Message) string {
			return m.OwnerLink
		}))

		return commentsMsg{messages: allMessages, owners: owners}
	}
}

// --- Helpers ---

// uniqueLinks extracts unique non-empty link strings from a slice.
func uniqueLinks[T any](items []T, getLink func(T) string) []string {
	seen := make(map[string]bool)
	var links []string
	for _, item := range items {
		link := getLink(item)
		if link != "" && !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}
	return links
}

// resolvePersonLinks fetches Person display names for a list of API links.
func resolvePersonLinks(client *launchpad.Client, links []string) map[string]string {
	result := make(map[string]string)
	for _, link := range links {
		req, err := http.NewRequest(http.MethodGet, link, nil)
		if err != nil {
			result[link] = link
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			result[link] = link
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK {
			result[link] = link
			continue
		}

		var person launchpad.Person
		if err := json.Unmarshal(body, &person); err != nil {
			result[link] = link
			continue
		}

		result[link] = fmt.Sprintf("%s (%s)", person.DisplayName, person.Name)
	}
	return result
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Only propagate to the active sub-model.
		var cmd tea.Cmd
		switch m.view {
		case viewList:
			m.list, cmd = m.list.Update(msg)
		case viewDetail:
			m.detail, cmd = m.detail.Update(msg)
		}
		return m, cmd

	case tea.KeyMsg:
		// Global quit.
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case searchResultsMsg:
		m.loading = false
		if msg.err != nil {
			m.status = errorStyle.Render(fmt.Sprintf("Error: %v", msg.err))
			m.view = viewSearch
			return m, nil
		}
		if len(msg.tasks) == 0 {
			m.status = "No results found."
			m.view = viewSearch
			return m, nil
		}
		m.status = ""
		m.list = newListModel(msg.tasks, msg.query, m.width, m.height)
		m.view = viewList
		return m, nil

	case bugDetailMsg:
		if msg.err != nil {
			m.loading = false
			m.status = errorStyle.Render(fmt.Sprintf("Error: %v", msg.err))
			m.view = m.prevView
			return m, nil
		}
		m.status = ""
		m.detail = newDetailModel(m.width, m.height)
		m.detail.setBug(msg.bug)
		m.view = viewDetail

		// Kick off tasks and comments fetch.
		var cmds []tea.Cmd
		if msg.bug.BugTasksCollectionLink != "" {
			cmds = append(cmds, fetchBugTasksCmd(m.client, msg.bug.BugTasksCollectionLink))
		}
		if msg.bug.MessagesCollectionLink != "" {
			cmds = append(cmds, fetchCommentsCmd(m.client, msg.bug.MessagesCollectionLink))
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		m.loading = false
		return m, nil

	case bugTasksMsg:
		if msg.err == nil {
			m.detail.setTasks(msg.tasks, msg.assignees)
		}
		// Check if comments are also done.
		if m.detail.messages != nil || msg.err != nil {
			m.loading = false
		}
		return m, nil

	case commentsMsg:
		if msg.err == nil {
			m.detail.setComments(msg.messages, msg.owners)
		}
		m.loading = false
		return m, nil
	}

	// Dispatch to current view.
	switch m.view {
	case viewSearch:
		return m.updateSearch(msg)
	case viewList:
		return m.updateList(msg)
	case viewDetail:
		return m.updateDetail(msg)
	}

	return m, nil
}

func (m model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if errMsg := m.search.validate(); errMsg != "" {
				m.status = errorStyle.Render(errMsg)
				return m, nil
			}
			m.status = m.search.statusLine()
			m.loading = true

			if bugID, ok := m.search.isDirectBugID(); ok {
				m.prevView = viewSearch
				return m, fetchBugCmd(m.client, bugID)
			}
			return m, searchBugsCmd(m.client, m.search.project(), m.search.searchText())
		}
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.list.SettingFilter() {
			break // let the list handle filter input
		}
		switch msg.String() {
		case "enter":
			bugID := m.list.selectedBugID()
			if bugID > 0 {
				m.loading = true
				m.status = fmt.Sprintf("Fetching bug #%d...", bugID)
				m.prevView = viewList
				return m, fetchBugCmd(m.client, bugID)
			}
		case "esc":
			m.view = viewSearch
			m.status = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.detail.reset()
			m.view = m.prevView
			m.status = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

// --- View ---

func (m model) View() string {
	var b strings.Builder

	switch m.view {
	case viewSearch:
		b.WriteString(m.search.View())
	case viewList:
		b.WriteString(m.list.View())
	case viewDetail:
		b.WriteString(m.detail.View())
	}

	if m.status != "" {
		b.WriteString("\n")
		if m.loading {
			b.WriteString(m.spinner.View() + " ")
		}
		b.WriteString(m.status)
	}

	return b.String()
}
