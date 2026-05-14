package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Update a resource",
}

var setBugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Update bug fields",
	RunE:  runSetBug,
}

var (
	setBugID         int
	setBugProject    string
	setBugTitle      string
	setBugAssignee   string
	setBugStatus     string
	setBugImportance string
	setBugTask       string
)

func init() {
	setBugCmd.Flags().IntVar(&setBugID, "id", 0, "Bug ID (required)")
	setBugCmd.MarkFlagRequired("id")
	setBugCmd.Flags().StringVar(&setBugProject, "project", "", "Project or distribution name (required)")
	setBugCmd.MarkFlagRequired("project")
	setBugCmd.Flags().StringVar(&setBugTitle, "title", "", "New bug title")
	setBugCmd.Flags().StringVar(&setBugAssignee, "assignee", "", "Assignee username (use \"\" to unassign)")
	setBugCmd.Flags().StringVar(&setBugStatus, "status", "", "Bug task status (e.g. New, Confirmed, \"In Progress\", \"Fix Committed\")")
	setBugCmd.Flags().StringVar(&setBugImportance, "importance", "", "Bug task importance (e.g. Critical, High, Medium)")
	setBugCmd.Flags().StringVar(&setBugTask, "task", "", "Bug task target name (required with --assignee, --status, --importance)")

	setCmd.AddCommand(setBugCmd)
	rootCmd.AddCommand(setCmd)
}

func runSetBug(cmd *cobra.Command, args []string) error {
	assigneeSet := cmd.Flags().Changed("assignee")
	statusSet := cmd.Flags().Changed("status")
	importanceSet := cmd.Flags().Changed("importance")

	// Require at least one field to update.
	if setBugTitle == "" && !assigneeSet && !statusSet && !importanceSet {
		return fmt.Errorf("at least one field flag is required (e.g. --title, --assignee, --status, --importance)")
	}

	// --assignee, --status, and --importance require --task.
	if (assigneeSet || statusSet || importanceSet) && setBugTask == "" {
		return fmt.Errorf("--task is required when using --assignee, --status, or --importance")
	}

	// Validate --status value locally.
	if statusSet {
		if !validBugTaskStatus(setBugStatus) {
			return fmt.Errorf("invalid status %q; valid values: %s", setBugStatus, validBugTaskStatusList())
		}
	}

	// Validate --importance value locally.
	if importanceSet {
		if !validBugTaskImportance(setBugImportance) {
			return fmt.Errorf("invalid importance %q; valid values: %s", setBugImportance, validBugTaskImportanceList())
		}
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	// Gate: fetch bug and tasks, verify the bug has a task matching --project.
	bug, err := client.GetBug(setBugID)
	if err != nil {
		return err
	}

	tasks, err := bug.GetTasks()
	if err != nil {
		return fmt.Errorf("fetching tasks: %w", err)
	}

	if _, err := findTaskByTarget(tasks, setBugProject); err != nil {
		return err
	}

	if setBugTitle != "" {
		if err := bug.SetTitle(setBugTitle); err != nil {
			return err
		}
	}

	if assigneeSet || statusSet || importanceSet {
		task, err := findTaskByTarget(tasks, setBugTask)
		if err != nil {
			return err
		}

		if assigneeSet {
			if err := task.SetAssignee(setBugAssignee); err != nil {
				return err
			}
			if setBugAssignee == "" {
				fmt.Printf("Unassigned task %q on bug #%d\n", setBugTask, setBugID)
			} else {
				fmt.Printf("Assigned %q to task %q on bug #%d\n", setBugAssignee, setBugTask, setBugID)
			}
		}

		if statusSet {
			if err := task.SetStatus(setBugStatus); err != nil {
				return err
			}
			fmt.Printf("Set status %q on task %q for bug #%d\n", setBugStatus, setBugTask, setBugID)
		}

		if importanceSet {
			if err := task.SetImportance(setBugImportance); err != nil {
				return err
			}
			fmt.Printf("Set importance %q on task %q for bug #%d\n", setBugImportance, setBugTask, setBugID)
		}
	}

	// Display updated bug.
	return showBug(client, setBugID, false)
}
