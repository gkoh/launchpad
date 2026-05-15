package main

import (
	"fmt"
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
	client, err := newClient()
	if err != nil {
		return err
	}

	var tags []string
	if searchBugTags != "" {
		for _, tag := range strings.Split(searchBugTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	tasks, err := client.SearchTasks(searchBugProject, &launchpad.SearchTasksOptions{
		SearchText: strings.Join(args, " "),
		Status:     searchBugStatus,
		Importance: searchBugImportance,
		Assignee:   searchBugAssignee,
		Tags:       tags,
		PageSize:   searchBugLimit,
	})
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("# Search Results (%d)\n", len(tasks))
	for _, task := range tasks {
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
