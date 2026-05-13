package launchpad

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBranchTypeJSON(t *testing.T) {
	tests := []struct {
		val  BranchType
		want string
	}{
		{BranchTypeHosted, `"Hosted"`},
		{BranchTypeMirrored, `"Mirrored"`},
		{BranchTypeImported, `"Imported"`},
		{BranchTypeRemote, `"Remote"`},
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

		var got BranchType
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestLifecycleStatusJSON(t *testing.T) {
	tests := []struct {
		val  LifecycleStatus
		want string
	}{
		{LifecycleExperimental, `"Experimental"`},
		{LifecycleDevelopment, `"Development"`},
		{LifecycleMature, `"Mature"`},
		{LifecycleMerged, `"Merged"`},
		{LifecycleAbandoned, `"Abandoned"`},
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

		var got LifecycleStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestBranchJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	earlier := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	orig := Branch{
		BranchFormat:                    "Bazaar Branch Format 7 (needs bzr 1.6)",
		BranchType:                      BranchTypeHosted,
		BzrIdentity:                     "lp:~user/project/trunk",
		CodeImportLink:                  NewLink("https://api.launchpad.net/devel/~user/project/trunk/+code-import"),
		ControlFormat:                   "Bazaar-NG meta directory, format 1",
		DateCreated:                     &earlier,
		DateLastModified:                &now,
		DependentBranchesCollectionLink: NewLink("https://api.launchpad.net/devel/~user/project/trunk/dependent_branches"),
		Description:                     "Main development branch",
		DisplayName:                     "lp:~user/project/trunk",
		ExplicitlyPrivate:               false,
		HTTPEtag:                        "\"abc123\"",
		InformationType:                 InformationPublic,
		LandingCandidatesCollectionLink: NewLink("https://api.launchpad.net/devel/~user/project/trunk/landing_candidates"),
		LandingTargetsCollectionLink:    NewLink("https://api.launchpad.net/devel/~user/project/trunk/landing_targets"),
		LastMirrorAttempt:               nil,
		LastMirrored:                    nil,
		LastScanned:                     &now,
		LastScannedID:                   "rev-42",
		LifecycleStatus:                 LifecycleDevelopment,
		LinkedBugsCollectionLink:        NewLink("https://api.launchpad.net/devel/~user/project/trunk/linked_bugs"),
		MirrorStatusMessage:             "",
		Name:                            "trunk",
		OwnerLink:                       NewLink("https://api.launchpad.net/devel/~user"),
		Private:                         false,
		ProjectLink:                     NewLink("https://api.launchpad.net/devel/project"),
		RecipesCollectionLink:           NewLink("https://api.launchpad.net/devel/~user/project/trunk/recipes"),
		RegistrantLink:                  NewLink("https://api.launchpad.net/devel/~user"),
		RepositoryFormat:                "Bazaar repository format 2a (needs bzr 1.16 or later)",
		ResourceTypeLink:                NewLink("https://api.launchpad.net/devel/#branch"),
		ReviewerLink:                    NewLink("https://api.launchpad.net/devel/~reviewer"),
		RevisionCount:                   42,
		SelfLink:                        NewLink("https://api.launchpad.net/devel/~user/project/trunk"),
		SpecLinksCollectionLink:         NewLink("https://api.launchpad.net/devel/~user/project/trunk/spec_links"),
		SubscribersCollectionLink:       NewLink("https://api.launchpad.net/devel/~user/project/trunk/subscribers"),
		SubscriptionsCollectionLink:     NewLink("https://api.launchpad.net/devel/~user/project/trunk/subscriptions"),
		UniqueName:                      "~user/project/trunk",
		URL:                             "",
		WebLink:                         NewLink("https://code.launchpad.net/~user/project/trunk"),
		WebhooksCollectionLink:          NewLink("https://api.launchpad.net/devel/~user/project/trunk/webhooks"),
		Whiteboard:                      "Some notes",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Branch
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if orig.Name != got.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if orig.WebLink.String() != got.WebLink.String() {
		t.Errorf("WebLink = %q, want %q", got.WebLink, orig.WebLink)
	}
	if orig.OwnerLink.String() != got.OwnerLink.String() {
		t.Errorf("OwnerLink = %q, want %q", got.OwnerLink, orig.OwnerLink)
	}
	if orig.RevisionCount != got.RevisionCount {
		t.Errorf("RevisionCount = %d, want %d", got.RevisionCount, orig.RevisionCount)
	}
}

func TestBranchJSONNulls(t *testing.T) {
	input := `{
		"branch_format": "",
		"branch_type": "Hosted",
		"bzr_identity": "",
		"control_format": "",
		"dependent_branches_collection_link": "",
		"description": "",
		"display_name": "",
		"explicitly_private": false,
		"http_etag": "",
		"information_type": "Public",
		"landing_candidates_collection_link": "",
		"landing_targets_collection_link": "",
		"last_scanned_id": "",
		"lifecycle_status": "Development",
		"linked_bugs_collection_link": "",
		"mirror_status_message": "",
		"name": "trunk",
		"owner_link": "",
		"private": false,
		"project_link": "",
		"recipes_collection_link": "",
		"registrant_link": "",
		"repository_format": "",
		"resource_type_link": "",
		"revision_count": 0,
		"self_link": "",
		"spec_links_collection_link": "",
		"subscribers_collection_link": "",
		"subscriptions_collection_link": "",
		"unique_name": "",
		"web_link": "",
		"webhooks_collection_link": ""
	}`

	var got Branch
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Name != "trunk" {
		t.Errorf("Name = %q, want %q", got.Name, "trunk")
	}
	if got.DateCreated != nil {
		t.Errorf("DateCreated = %v, want nil", got.DateCreated)
	}
	if got.DateLastModified != nil {
		t.Errorf("DateLastModified = %v, want nil", got.DateLastModified)
	}
	if got.LastMirrorAttempt != nil {
		t.Errorf("LastMirrorAttempt = %v, want nil", got.LastMirrorAttempt)
	}
	if got.LastMirrored != nil {
		t.Errorf("LastMirrored = %v, want nil", got.LastMirrored)
	}
	if got.LastScanned != nil {
		t.Errorf("LastScanned = %v, want nil", got.LastScanned)
	}
	if !got.CodeImportLink.IsZero() {
		t.Errorf("CodeImportLink = %q, want zero", got.CodeImportLink)
	}
	if !got.ReviewerLink.IsZero() {
		t.Errorf("ReviewerLink = %q, want zero", got.ReviewerLink)
	}
}

func TestBranchCollectionJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := BranchCollection{
		CollectionMeta: CollectionMeta{
			TotalSize:          2,
			Start:              0,
			NextCollectionLink: NewLink("https://api.launchpad.net/devel/branches?ws.start=2"),
		},
		Entries: []Branch{
			{
				Name:            "trunk",
				BranchType:      BranchTypeHosted,
				LifecycleStatus: LifecycleDevelopment,
				InformationType: InformationPublic,
				DateCreated:     &now,
				RevisionCount:   10,
				SelfLink:        NewLink("https://api.launchpad.net/devel/~user/project/trunk"),
				WebLink:         NewLink("https://code.launchpad.net/~user/project/trunk"),
			},
			{
				Name:            "feature",
				BranchType:      BranchTypeHosted,
				LifecycleStatus: LifecycleExperimental,
				InformationType: InformationPrivate,
				DateCreated:     &now,
				RevisionCount:   3,
				Private:         true,
				SelfLink:        NewLink("https://api.launchpad.net/devel/~user/project/feature"),
				WebLink:         NewLink("https://code.launchpad.net/~user/project/feature"),
			},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got BranchCollection
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", got.TotalSize)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].BranchType != BranchTypeHosted {
		t.Errorf("Entries[0].BranchType = %q, want %q", got.Entries[0].BranchType, BranchTypeHosted)
	}
	if got.Entries[1].InformationType != InformationPrivate {
		t.Errorf("Entries[1].InformationType = %q, want %q", got.Entries[1].InformationType, InformationPrivate)
	}
}
