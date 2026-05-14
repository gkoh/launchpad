package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gkoh/launchpad"
)

// statusContentReturned is Launchpad's non-standard HTTP 209 response
// indicating a successful PATCH with content returned.
const statusContentReturned = 209

// maxDescriptionLength is the maximum number of characters to display
// for a bug description before truncating.
const maxDescriptionLength = 200

// maxCommentLength is the maximum number of characters to display
// for a comment before truncating.
const maxCommentLength = 500

// findTaskByTarget returns the bug task whose BugTargetName matches the given
// name exactly. If no match is found, it returns an error listing available
// target names.
func findTaskByTarget(tasks []launchpad.BugTask, targetName string) (*launchpad.BugTask, error) {
	for i := range tasks {
		if tasks[i].BugTargetName == targetName {
			return &tasks[i], nil
		}
	}
	var names []string
	for _, t := range tasks {
		names = append(names, t.BugTargetName)
	}
	return nil, fmt.Errorf("no task found for target %q; available targets: %s", targetName, strings.Join(names, ", "))
}

// validBugTaskStatus returns true if s is a valid BugTaskStatus value.
func validBugTaskStatus(s string) bool {
	switch launchpad.BugTaskStatus(s) {
	case launchpad.BugTaskStatusNew,
		launchpad.BugTaskStatusIncomplete,
		launchpad.BugTaskStatusOpinion,
		launchpad.BugTaskStatusInvalid,
		launchpad.BugTaskStatusWontFix,
		launchpad.BugTaskStatusExpired,
		launchpad.BugTaskStatusConfirmed,
		launchpad.BugTaskStatusTriaged,
		launchpad.BugTaskStatusInProgress,
		launchpad.BugTaskStatusDeferred,
		launchpad.BugTaskStatusFixCommitted,
		launchpad.BugTaskStatusFixReleased,
		launchpad.BugTaskStatusDoesNotExist,
		launchpad.BugTaskStatusUnknown:
		return true
	}
	return false
}

// validBugTaskStatusList returns a comma-separated string of all valid
// BugTaskStatus values for use in error messages.
func validBugTaskStatusList() string {
	return strings.Join([]string{
		string(launchpad.BugTaskStatusNew),
		string(launchpad.BugTaskStatusIncomplete),
		string(launchpad.BugTaskStatusOpinion),
		string(launchpad.BugTaskStatusInvalid),
		string(launchpad.BugTaskStatusWontFix),
		string(launchpad.BugTaskStatusExpired),
		string(launchpad.BugTaskStatusConfirmed),
		string(launchpad.BugTaskStatusTriaged),
		string(launchpad.BugTaskStatusInProgress),
		string(launchpad.BugTaskStatusDeferred),
		string(launchpad.BugTaskStatusFixCommitted),
		string(launchpad.BugTaskStatusFixReleased),
		string(launchpad.BugTaskStatusDoesNotExist),
		string(launchpad.BugTaskStatusUnknown),
	}, ", ")
}

// validBugTaskImportance returns true if s is a valid BugTaskImportance value.
func validBugTaskImportance(s string) bool {
	switch launchpad.BugTaskImportance(s) {
	case launchpad.BugTaskImportanceUnknown,
		launchpad.BugTaskImportanceUndecided,
		launchpad.BugTaskImportanceCritical,
		launchpad.BugTaskImportanceHigh,
		launchpad.BugTaskImportanceMedium,
		launchpad.BugTaskImportanceLow,
		launchpad.BugTaskImportanceWishlist:
		return true
	}
	return false
}

// validBugTaskImportanceList returns a comma-separated string of all valid
// BugTaskImportance values for use in error messages.
func validBugTaskImportanceList() string {
	return strings.Join([]string{
		string(launchpad.BugTaskImportanceUnknown),
		string(launchpad.BugTaskImportanceUndecided),
		string(launchpad.BugTaskImportanceCritical),
		string(launchpad.BugTaskImportanceHigh),
		string(launchpad.BugTaskImportanceMedium),
		string(launchpad.BugTaskImportanceLow),
		string(launchpad.BugTaskImportanceWishlist),
	}, ", ")
}

// fetchBugTasks fetches the bug task collection from the given URL.
func fetchBugTasks(client *launchpad.Client, url string) ([]launchpad.BugTask, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bug tasks returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var collection launchpad.BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, err
	}

	return collection.Entries, nil
}

// resolveAssignees returns a map from assignee link URL to "DisplayName (name)"
// for all unique assignees across the given tasks.
func resolveAssignees(client *launchpad.Client, tasks []launchpad.BugTask) map[string]string {
	var links []launchpad.Link
	for _, task := range tasks {
		links = append(links, task.AssigneeLink)
	}
	return client.ResolvePersonLinks(links)
}

// fetchAllMessages fetches all messages for a bug, following pagination links.
func fetchAllMessages(client *launchpad.Client, url string) ([]launchpad.Message, error) {
	var all []launchpad.Message

	for url != "" {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return all, err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return all, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return all, err
		}

		if resp.StatusCode != http.StatusOK {
			return all, fmt.Errorf("messages returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var collection launchpad.MessageCollection
		if err := json.Unmarshal(body, &collection); err != nil {
			return all, err
		}

		all = append(all, collection.Entries...)
		url = collection.NextCollectionLink.String()
	}

	return all, nil
}

// resolveOwners returns a map from owner link URL to "DisplayName (name)"
// for all unique message owners.
func resolveOwners(client *launchpad.Client, messages []launchpad.Message) map[string]string {
	var links []launchpad.Link
	for _, msg := range messages {
		links = append(links, msg.OwnerLink)
	}
	return client.ResolvePersonLinks(links)
}
