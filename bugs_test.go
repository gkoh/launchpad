package launchpad

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLockStatusJSON(t *testing.T) {
	tests := []struct {
		val  LockStatus
		want string
	}{
		{LockStatusUnlocked, `"Unlocked"`},
		{LockStatusCommentOnly, `"Comment-only"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.val)
		if err != nil {
			t.Errorf("Marshal(%q): %v", tt.val, err)
			continue
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%q) = %s, want %s", tt.val, data, tt.want)
		}

		var got LockStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestBugJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	earlier := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	orig := Bug{
		ActivityCollectionLink:               NewLink("https://api.launchpad.net/devel/bugs/1/activity"),
		AttachmentsCollectionLink:            NewLink("https://api.launchpad.net/devel/bugs/1/attachments"),
		BugTasksCollectionLink:               NewLink("https://api.launchpad.net/devel/bugs/1/bug_tasks"),
		BugWatchesCollectionLink:             NewLink("https://api.launchpad.net/devel/bugs/1/bug_watches"),
		CVEsCollectionLink:                   NewLink("https://api.launchpad.net/devel/bugs/1/cves"),
		DateCreated:                          &earlier,
		DateLastMessage:                      &now,
		DateLastUpdated:                      &now,
		DateMadePrivate:                      nil,
		Description:                          "A detailed bug description.",
		DuplicatesCollectionLink:             NewLink("https://api.launchpad.net/devel/bugs/1/duplicates"),
		Heat:                                 42,
		HTTPEtag:                             "\"etag123\"",
		ID:                                   1,
		InformationType:                      InformationPublic,
		LatestPatchUploaded:                  &now,
		LinkedBranchesCollectionLink:         NewLink("https://api.launchpad.net/devel/bugs/1/linked_branches"),
		LinkedMergeProposalsCollectionLink:   NewLink("https://api.launchpad.net/devel/bugs/1/linked_merge_proposals"),
		LockReason:                           "",
		LockStatus:                           LockStatusUnlocked,
		MessageCount:                         5,
		MessagesCollectionLink:               NewLink("https://api.launchpad.net/devel/bugs/1/messages"),
		Name:                                 "test-bug",
		NumberOfDuplicates:                   2,
		OtherUsersAffectedCountWithDupes:     10,
		OwnerLink:                            NewLink("https://api.launchpad.net/devel/~user"),
		Private:                              false,
		ResourceTypeLink:                     NewLink("https://api.launchpad.net/devel/#bug"),
		SecurityRelated:                      false,
		SelfLink:                             NewLink("https://api.launchpad.net/devel/bugs/1"),
		SubscriptionsCollectionLink:          NewLink("https://api.launchpad.net/devel/bugs/1/subscriptions"),
		Tags:                                 []string{"kernel", "regression"},
		Title:                                "Test bug title",
		UsersAffectedCollectionLink:          NewLink("https://api.launchpad.net/devel/bugs/1/users_affected"),
		UsersAffectedCount:                   3,
		UsersAffectedCountWithDupes:          5,
		UsersAffectedWithDupesCollectionLink: NewLink("https://api.launchpad.net/devel/bugs/1/users_affected_with_dupes"),
		UsersUnaffectedCollectionLink:        NewLink("https://api.launchpad.net/devel/bugs/1/users_unaffected"),
		UsersUnaffectedCount:                 1,
		VulnerabilitiesCollectionLink:        NewLink("https://api.launchpad.net/devel/bugs/1/vulnerabilities"),
		WebLink:                              NewLink("https://bugs.launchpad.net/bugs/1"),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Bug
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if orig.ID != got.ID {
		t.Errorf("ID = %d, want %d", got.ID, orig.ID)
	}
	if orig.WebLink.String() != got.WebLink.String() {
		t.Errorf("WebLink = %q, want %q", got.WebLink, orig.WebLink)
	}
	if orig.BugTasksCollectionLink.String() != got.BugTasksCollectionLink.String() {
		t.Errorf("BugTasksCollectionLink = %q, want %q", got.BugTasksCollectionLink, orig.BugTasksCollectionLink)
	}
	if orig.Title != got.Title {
		t.Errorf("Title = %q, want %q", got.Title, orig.Title)
	}
}

func TestBugJSONNulls(t *testing.T) {
	input := `{
		"activity_collection_link": "",
		"attachments_collection_link": "",
		"bug_tasks_collection_link": "",
		"bug_watches_collection_link": "",
		"cves_collection_link": "",
		"description": "",
		"duplicates_collection_link": "",
		"heat": 0,
		"http_etag": "",
		"id": 99,
		"information_type": "Public",
		"linked_branches_collection_link": "",
		"linked_merge_proposals_collection_link": "",
		"lock_status": "Unlocked",
		"message_count": 0,
		"messages_collection_link": "",
		"number_of_duplicates": 0,
		"other_users_affected_count_with_dupes": 0,
		"owner_link": "",
		"private": false,
		"resource_type_link": "",
		"security_related": false,
		"self_link": "",
		"subscriptions_collection_link": "",
		"tags": [],
		"title": "Minimal bug",
		"users_affected_collection_link": "",
		"users_affected_count": 0,
		"users_affected_count_with_dupes": 0,
		"users_affected_with_dupes_collection_link": "",
		"users_unaffected_collection_link": "",
		"users_unaffected_count": 0,
		"vulnerabilities_collection_link": "",
		"web_link": ""
	}`

	var got Bug
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.ID != 99 {
		t.Errorf("ID = %d, want 99", got.ID)
	}
	if got.Title != "Minimal bug" {
		t.Errorf("Title = %q, want %q", got.Title, "Minimal bug")
	}
	if got.DateCreated != nil {
		t.Errorf("DateCreated = %v, want nil", got.DateCreated)
	}
	if got.DateLastMessage != nil {
		t.Errorf("DateLastMessage = %v, want nil", got.DateLastMessage)
	}
	if got.DateLastUpdated != nil {
		t.Errorf("DateLastUpdated = %v, want nil", got.DateLastUpdated)
	}
	if got.DateMadePrivate != nil {
		t.Errorf("DateMadePrivate = %v, want nil", got.DateMadePrivate)
	}
	if got.LatestPatchUploaded != nil {
		t.Errorf("LatestPatchUploaded = %v, want nil", got.LatestPatchUploaded)
	}
	if len(got.Tags) != 0 {
		t.Errorf("Tags = %v, want empty", got.Tags)
	}
}

func TestBugCollectionJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := BugCollection{
		CollectionMeta: CollectionMeta{
			TotalSize:          2,
			Start:              0,
			NextCollectionLink: NewLink("https://api.launchpad.net/devel/bugs?ws.start=2"),
		},
		Entries: []Bug{
			{
				ID:              1,
				Title:           "First bug",
				InformationType: InformationPublic,
				LockStatus:      LockStatusUnlocked,
				DateCreated:     &now,
				Heat:            10,
				Tags:            []string{"ui"},
				WebLink:         NewLink("https://bugs.launchpad.net/bugs/1"),
			},
			{
				ID:              2,
				Title:           "Second bug",
				InformationType: InformationPrivate,
				LockStatus:      LockStatusCommentOnly,
				DateCreated:     &now,
				Heat:            99,
				Private:         true,
				Tags:            []string{"security", "critical"},
				WebLink:         NewLink("https://bugs.launchpad.net/bugs/2"),
			},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got BugCollection
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", got.TotalSize)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].LockStatus != LockStatusUnlocked {
		t.Errorf("Entries[0].LockStatus = %q", got.Entries[0].LockStatus)
	}
	if got.Entries[1].LockStatus != LockStatusCommentOnly {
		t.Errorf("Entries[1].LockStatus = %q", got.Entries[1].LockStatus)
	}
}

func TestBugTaskStatusJSON(t *testing.T) {
	tests := []struct {
		val  BugTaskStatus
		want string
	}{
		{BugTaskStatusNew, `"New"`},
		{BugTaskStatusIncomplete, `"Incomplete"`},
		{BugTaskStatusOpinion, `"Opinion"`},
		{BugTaskStatusInvalid, `"Invalid"`},
		{BugTaskStatusWontFix, `"Won't Fix"`},
		{BugTaskStatusExpired, `"Expired"`},
		{BugTaskStatusConfirmed, `"Confirmed"`},
		{BugTaskStatusTriaged, `"Triaged"`},
		{BugTaskStatusInProgress, `"In Progress"`},
		{BugTaskStatusDeferred, `"Deferred"`},
		{BugTaskStatusFixCommitted, `"Fix Committed"`},
		{BugTaskStatusFixReleased, `"Fix Released"`},
		{BugTaskStatusDoesNotExist, `"Does Not Exist"`},
		{BugTaskStatusUnknown, `"Unknown"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.val)
		if err != nil {
			t.Errorf("Marshal(%q): %v", tt.val, err)
			continue
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%q) = %s, want %s", tt.val, data, tt.want)
		}

		var got BugTaskStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestBugTaskImportanceJSON(t *testing.T) {
	tests := []struct {
		val  BugTaskImportance
		want string
	}{
		{BugTaskImportanceUnknown, `"Unknown"`},
		{BugTaskImportanceUndecided, `"Undecided"`},
		{BugTaskImportanceCritical, `"Critical"`},
		{BugTaskImportanceHigh, `"High"`},
		{BugTaskImportanceMedium, `"Medium"`},
		{BugTaskImportanceLow, `"Low"`},
		{BugTaskImportanceWishlist, `"Wishlist"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.val)
		if err != nil {
			t.Errorf("Marshal(%q): %v", tt.val, err)
			continue
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%q) = %s, want %s", tt.val, data, tt.want)
		}

		var got BugTaskImportance
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestBugTaskJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := BugTask{
		AssigneeLink:               NewLink("https://api.launchpad.net/devel/~assignee"),
		BugLink:                    NewLink("https://api.launchpad.net/devel/bugs/1"),
		BugTargetDisplayName:       "linux (Ubuntu)",
		BugTargetName:              "linux (Ubuntu)",
		DateAssigned:               &now,
		DateClosed:                 nil,
		DateConfirmed:              &now,
		DateCreated:                &now,
		DateDeferred:               nil,
		DateFixCommitted:           nil,
		DateFixReleased:            nil,
		DateInProgress:             nil,
		DateIncomplete:             nil,
		DateLeftClosed:             nil,
		DateLeftNew:                &now,
		DateTriaged:                &now,
		HTTPEtag:                   "\"etag456\"",
		Importance:                 BugTaskImportanceHigh,
		ImportanceExplanation:      "Affects many users",
		IsComplete:                 false,
		MilestoneLink:              NewLink("https://api.launchpad.net/devel/ubuntu/+milestone/noble"),
		OwnerLink:                  NewLink("https://api.launchpad.net/devel/~owner"),
		RelatedTasksCollectionLink: NewLink("https://api.launchpad.net/devel/ubuntu/+source/linux/+bug/1/related_tasks"),
		ResourceTypeLink:           NewLink("https://api.launchpad.net/devel/#bug_task"),
		SelfLink:                   NewLink("https://api.launchpad.net/devel/ubuntu/+source/linux/+bug/1"),
		Status:                     BugTaskStatusConfirmed,
		StatusExplanation:          "Confirmed by multiple reporters",
		TargetLink:                 NewLink("https://api.launchpad.net/devel/ubuntu/+source/linux"),
		Title:                      "Bug #1 in linux (Ubuntu): \"Test bug\"",
		WebLink:                    NewLink("https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1"),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got BugTask
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if orig.AssigneeLink.String() != got.AssigneeLink.String() {
		t.Errorf("AssigneeLink = %q, want %q", got.AssigneeLink, orig.AssigneeLink)
	}
	if orig.BugLink.String() != got.BugLink.String() {
		t.Errorf("BugLink = %q, want %q", got.BugLink, orig.BugLink)
	}
	if orig.Title != got.Title {
		t.Errorf("Title = %q, want %q", got.Title, orig.Title)
	}
	if orig.Status != got.Status {
		t.Errorf("Status = %q, want %q", got.Status, orig.Status)
	}
}

func TestBugTaskJSONNulls(t *testing.T) {
	input := `{
		"bug_link": "https://api.launchpad.net/devel/bugs/1",
		"bug_target_display_name": "linux (Ubuntu)",
		"bug_target_name": "linux (Ubuntu)",
		"http_etag": "",
		"importance": "Undecided",
		"is_complete": false,
		"owner_link": "",
		"related_tasks_collection_link": "",
		"resource_type_link": "",
		"self_link": "",
		"status": "New",
		"target_link": "",
		"title": "Bug #1",
		"web_link": ""
	}`

	var got BugTask
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !got.AssigneeLink.IsZero() {
		t.Errorf("AssigneeLink = %q, want zero", got.AssigneeLink)
	}
	if got.DateAssigned != nil {
		t.Errorf("DateAssigned = %v, want nil", got.DateAssigned)
	}
	if got.DateCreated != nil {
		t.Errorf("DateCreated = %v, want nil", got.DateCreated)
	}
	if !got.MilestoneLink.IsZero() {
		t.Errorf("MilestoneLink = %q, want zero", got.MilestoneLink)
	}
	if got.Status != BugTaskStatusNew {
		t.Errorf("Status = %q, want %q", got.Status, BugTaskStatusNew)
	}
	if got.Importance != BugTaskImportanceUndecided {
		t.Errorf("Importance = %q, want %q", got.Importance, BugTaskImportanceUndecided)
	}
}

func TestBugTaskCollectionJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := BugTaskCollection{
		CollectionMeta: CollectionMeta{
			TotalSize: 2,
			Start:     0,
		},
		Entries: []BugTask{
			{
				BugLink:              NewLink("https://api.launchpad.net/devel/bugs/1"),
				BugTargetDisplayName: "linux (Ubuntu)",
				BugTargetName:        "linux (Ubuntu)",
				Status:               BugTaskStatusConfirmed,
				Importance:           BugTaskImportanceHigh,
				AssigneeLink:         NewLink("https://api.launchpad.net/devel/~user1"),
				DateCreated:          &now,
				Title:                "Bug #1",
				WebLink:              NewLink("https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1"),
			},
			{
				BugLink:              NewLink("https://api.launchpad.net/devel/bugs/1"),
				BugTargetDisplayName: "linux (Ubuntu Noble)",
				BugTargetName:        "linux (Ubuntu Noble)",
				Status:               BugTaskStatusNew,
				Importance:           BugTaskImportanceUndecided,
				DateCreated:          &now,
				Title:                "Bug #1",
				WebLink:              NewLink("https://bugs.launchpad.net/ubuntu/noble/+source/linux/+bug/1"),
			},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got BugTaskCollection
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(got.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].Status != BugTaskStatusConfirmed {
		t.Errorf("Entries[0].Status = %q", got.Entries[0].Status)
	}
	if got.Entries[0].AssigneeLink.String() != "https://api.launchpad.net/devel/~user1" {
		t.Errorf("Entries[0].AssigneeLink = %q", got.Entries[0].AssigneeLink)
	}
	if !got.Entries[1].AssigneeLink.IsZero() {
		t.Errorf("Entries[1].AssigneeLink = %q, want zero", got.Entries[1].AssigneeLink)
	}
}

func TestBugTaskBugID(t *testing.T) {
	task := &BugTask{
		BugLink: NewLink("https://api.launchpad.net/devel/bugs/12345"),
	}
	id, err := task.BugID()
	if err != nil {
		t.Fatal(err)
	}
	if id != 12345 {
		t.Errorf("BugID() = %d, want 12345", id)
	}
}

func TestBugTaskBugIDZeroLink(t *testing.T) {
	task := &BugTask{}
	_, err := task.BugID()
	if err == nil {
		t.Fatal("expected error for zero bug link, got nil")
	}
}
