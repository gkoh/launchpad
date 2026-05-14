package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gkoh/launchpad"
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
	resp, err := client.Get(fmt.Sprintf("/bugs/%d", setBugID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading bug: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var bug launchpad.Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return fmt.Errorf("parsing bug: %w", err)
	}

	if bug.BugTasksCollectionLink.IsZero() {
		return fmt.Errorf("bug #%d has no tasks", setBugID)
	}

	tasks, err := fetchBugTasks(client, bug.BugTasksCollectionLink.String())
	if err != nil {
		return fmt.Errorf("fetching tasks: %w", err)
	}

	if _, err := findTaskByTarget(tasks, setBugProject); err != nil {
		return err
	}

	if setBugTitle != "" {
		if err := updateBugTitle(client, setBugID, setBugTitle); err != nil {
			return err
		}
	}

	if assigneeSet {
		if err := updateBugTaskAssignee(client, setBugID, tasks, setBugTask, setBugAssignee); err != nil {
			return err
		}
	}

	if statusSet {
		if err := updateBugTaskStatus(client, setBugID, tasks, setBugTask, setBugStatus); err != nil {
			return err
		}
	}

	if importanceSet {
		if err := updateBugTaskImportance(client, setBugID, tasks, setBugTask, setBugImportance); err != nil {
			return err
		}
	}

	// Display updated bug.
	return showBug(client, setBugID, false)
}

// updateBugTitle sends a PATCH request to update a bug's title.
func updateBugTitle(client *launchpad.Client, bugID int, title string) error {
	payload, err := json.Marshal(map[string]string{"title": title})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	resp, err := client.Patch(fmt.Sprintf("/bugs/%d", bugID), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("updating bug title: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth --permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// updateBugTaskAssignee sets the assignee on a bug task matching the given
// target name. An empty username unassigns the task. The tasks slice must
// have been fetched beforehand.
func updateBugTaskAssignee(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, username string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	// Build the PATCH payload.
	var payload []byte
	if username == "" {
		// Unassign: send null.
		payload = []byte(`{"assignee_link":null}`)
	} else {
		assigneeLink := fmt.Sprintf("https://api.launchpad.net/devel/~%s", username)
		payload, err = json.Marshal(map[string]string{"assignee_link": assigneeLink})
		if err != nil {
			return fmt.Errorf("marshalling payload: %w", err)
		}
	}

	// PATCH the task.
	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	patchResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating assignee: %w", err)
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth --permission WRITE_PRIVATE')", bugID)
	}
	if patchResp.StatusCode != http.StatusOK && patchResp.StatusCode != statusContentReturned {
		patchBody, _ := io.ReadAll(patchResp.Body)
		return fmt.Errorf("API returned %d: %s", patchResp.StatusCode, strings.TrimSpace(string(patchBody)))
	}

	if username == "" {
		fmt.Printf("Unassigned task %q on bug #%d\n", targetName, bugID)
	} else {
		fmt.Printf("Assigned %q to task %q on bug #%d\n", username, targetName, bugID)
	}
	return nil
}

// updateBugTaskStatus sets the status on a bug task matching the given target
// name. The tasks slice must have been fetched beforehand.
func updateBugTaskStatus(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, status string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth --permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Printf("Set status %q on task %q for bug #%d\n", status, targetName, bugID)
	return nil
}

// updateBugTaskImportance sets the importance on a bug task matching the given
// target name. The tasks slice must have been fetched beforehand.
func updateBugTaskImportance(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, importance string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{"importance": importance})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating importance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth --permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Printf("Set importance %q on task %q for bug #%d\n", importance, targetName, bugID)
	return nil
}
