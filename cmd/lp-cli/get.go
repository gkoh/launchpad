package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gkoh/launchpad"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch and display a resource",
}

var getBugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Query and display a bug by ID",
	RunE:  runGetBug,
}

var (
	getBugID      int
	getBugVerbose bool
)

func init() {
	getBugCmd.Flags().IntVar(&getBugID, "id", 0, "Bug ID (required)")
	getBugCmd.MarkFlagRequired("id")
	getBugCmd.Flags().BoolVar(&getBugVerbose, "verbose", false, "Show all comments")

	getCmd.AddCommand(getBugCmd)
	rootCmd.AddCommand(getCmd)
}

func runGetBug(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	return showBug(client, getBugID, getBugVerbose)
}

// showBug fetches and displays a bug by ID, including its tasks and assignees.
// When verbose is true, all comments are fetched and displayed.
func showBug(client *launchpad.Client, bugID int, verbose bool) error {
	bug, err := client.GetBug(bugID)
	if err != nil {
		return err
	}

	// Display bug summary.
	fmt.Printf("# Bug #%d: %s\n\n", bug.ID, bug.Title)
	fmt.Printf("- **Info type:** %s\n", bug.InformationType)
	fmt.Printf("- **Lock status:** %s\n", bug.LockStatus)
	fmt.Printf("- **Heat:** %d\n", bug.Heat)
	if len(bug.Tags) > 0 {
		fmt.Printf("- **Tags:** %s\n", strings.Join(bug.Tags, ", "))
	}
	if bug.DateCreated != nil {
		fmt.Printf("- **Created:** %s\n", bug.DateCreated.Format("2006-01-02 15:04:05"))
	}
	if bug.DateLastUpdated != nil {
		fmt.Printf("- **Updated:** %s\n", bug.DateLastUpdated.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("- **Messages:** %d\n", bug.MessageCount)
	fmt.Printf("- **Web:** %s\n", bug.WebLink)

	if bug.Description != "" {
		desc := bug.Description
		if len(desc) > maxDescriptionLength {
			desc = desc[:maxDescriptionLength] + "..."
		}
		fmt.Printf("\n## Description\n\n%s\n", desc)
	}

	// Fetch bug tasks.
	if !bug.BugTasksCollectionLink.IsZero() {
		tasks, err := fetchBugTasks(client, bug.BugTasksCollectionLink.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: could not fetch bug tasks: %v\n", err)
		} else if len(tasks) > 0 {
			// Resolve assignees.
			assignees := resolveAssignees(client, tasks)

			fmt.Printf("\n## Tasks\n")
			for _, task := range tasks {
				fmt.Printf("\n### %s\n\n", task.BugTargetDisplayName)
				fmt.Printf("- **Status:** %s\n", task.Status)
				fmt.Printf("- **Importance:** %s\n", task.Importance)
				if name, ok := assignees[task.AssigneeLink.String()]; ok {
					fmt.Printf("- **Assignee:** %s\n", name)
				} else {
					fmt.Printf("- **Assignee:** unassigned\n")
				}
			}
		}
	}

	// Fetch and display comments when verbose.
	if verbose && !bug.MessagesCollectionLink.IsZero() {
		messages, err := fetchAllMessages(client, bug.MessagesCollectionLink.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: could not fetch comments: %v\n", err)
		} else if len(messages) > 0 {
			// Resolve owner display names.
			owners := resolveOwners(client, messages)

			fmt.Printf("\n## Comments (%d)\n", len(messages))
			for i, msg := range messages {
				owner := msg.OwnerLink.String()
				if name, ok := owners[msg.OwnerLink.String()]; ok {
					owner = name
				}
				date := "unknown"
				if msg.DateCreated != nil {
					date = msg.DateCreated.Format("2006-01-02 15:04:05")
				}
				fmt.Printf("\n### #%d by %s on %s\n\n", i+1, owner, date)
				content := msg.Content
				if len(content) > maxCommentLength {
					content = content[:maxCommentLength] + "..."
				}
				if content != "" {
					fmt.Printf("%s\n", content)
				}
			}
		}
	}

	return nil
}
