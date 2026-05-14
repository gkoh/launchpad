package main

import (
	"fmt"
	"strings"

	"github.com/gkoh/launchpad"
)

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
