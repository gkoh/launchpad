package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gkoh/launchpad"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for resources",
}

var searchBugCmd = &cobra.Command{
	Use:   "bug [text]",
	Short: "Search for bugs",
	Long:  "Search for bugs by project, status, importance, assignee, tags, and free text.",
	RunE:  runSearchBug,
}

var (
	searchBugProject    string
	searchBugStatus     string
	searchBugImportance string
	searchBugAssignee   string
	searchBugTags       string
	searchBugLimit      int
)

func init() {
	searchBugCmd.Flags().StringVar(&searchBugProject, "project", "", "Project or distribution name (required)")
	searchBugCmd.MarkFlagRequired("project")
	searchBugCmd.Flags().StringVar(&searchBugStatus, "status", "", "Filter by bug task status (e.g. New, Confirmed, In Progress)")
	searchBugCmd.Flags().StringVar(&searchBugImportance, "importance", "", "Filter by importance (e.g. Critical, High, Medium)")
	searchBugCmd.Flags().StringVar(&searchBugAssignee, "assignee", "", "Filter by assignee username")
	searchBugCmd.Flags().StringVar(&searchBugTags, "tags", "", "Filter by tags (comma-separated)")
	searchBugCmd.Flags().IntVar(&searchBugLimit, "limit", 10, "Maximum number of results")

	searchCmd.AddCommand(searchBugCmd)
	rootCmd.AddCommand(searchCmd)
}

func runSearchBug(cmd *cobra.Command, args []string) error {
	// Remaining positional arguments form the search text.
	searchText := strings.Join(args, " ")

	// Build the search URL with ws.op=searchTasks.
	params := url.Values{}
	params.Set("ws.op", "searchTasks")
	if searchText != "" {
		params.Set("search_text", searchText)
	}
	if searchBugStatus != "" {
		params.Set("status", searchBugStatus)
	}
	if searchBugImportance != "" {
		params.Set("importance", searchBugImportance)
	}
	if searchBugAssignee != "" {
		params.Set("assignee", fmt.Sprintf("https://api.launchpad.net/devel/~%s", searchBugAssignee))
	}
	if searchBugTags != "" {
		for _, tag := range strings.Split(searchBugTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				params.Add("tags", tag)
			}
		}
	}
	params.Set("ws.size", fmt.Sprintf("%d", searchBugLimit))

	path := fmt.Sprintf("/%s?%s", searchBugProject, params.Encode())

	client, err := newClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project %q not found", searchBugProject)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var collection launchpad.BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	if len(collection.Entries) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("# Search Results (%d)\n", len(collection.Entries))
	for _, task := range collection.Entries {
		fmt.Printf("\n## %s\n\n", task.Title)
		fmt.Printf("- **Target:** %s\n", task.BugTargetDisplayName)
		fmt.Printf("- **Status:** %s\n", task.Status)
		fmt.Printf("- **Importance:** %s\n", task.Importance)
		if !task.AssigneeLink.IsZero() {
			fmt.Printf("- **Assignee:** %s\n", task.AssigneeLink)
		}
		fmt.Printf("- **Web:** %s\n", task.WebLink)
	}
	return nil
}
