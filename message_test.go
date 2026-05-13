package launchpad

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMessageJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := Message{
		Content:          "This is a test comment on the bug.",
		DateCreated:      &now,
		HTTPEtag:         "\"etag123\"",
		OwnerLink:        "https://api.launchpad.net/devel/~test-user",
		ResourceTypeLink: "https://api.launchpad.net/devel/#message",
		SelfLink:         "https://api.launchpad.net/devel/bugs/1/+message/0",
		Subject:          "Bug #1: Test bug",
		WebLink:          "https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1/comments/0",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Message
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", got, orig)
	}
}

func TestMessageJSONNulls(t *testing.T) {
	input := `{
		"content": "A comment",
		"http_etag": "",
		"owner_link": "",
		"resource_type_link": "",
		"self_link": "",
		"subject": "",
		"web_link": ""
	}`

	var got Message
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Content != "A comment" {
		t.Errorf("Content = %q, want %q", got.Content, "A comment")
	}
	if got.DateCreated != nil {
		t.Errorf("DateCreated = %v, want nil", got.DateCreated)
	}
}

func TestMessageCollectionJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	later := time.Date(2025, 6, 16, 9, 30, 0, 0, time.UTC)

	orig := MessageCollection{
		CollectionMeta: CollectionMeta{
			TotalSize: 2,
			Start:     0,
		},
		Entries: []Message{
			{
				Content:     "Initial report of the bug.",
				DateCreated: &now,
				OwnerLink:   "https://api.launchpad.net/devel/~reporter",
				Subject:     "Bug #1: Test bug",
				WebLink:     "https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1/comments/0",
			},
			{
				Content:     "I can confirm this issue.",
				DateCreated: &later,
				OwnerLink:   "https://api.launchpad.net/devel/~confirmer",
				Subject:     "Re: Bug #1: Test bug",
				WebLink:     "https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1/comments/1",
			},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got MessageCollection
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", got, orig)
	}

	if len(got.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].Content != "Initial report of the bug." {
		t.Errorf("Entries[0].Content = %q", got.Entries[0].Content)
	}
	if got.Entries[1].OwnerLink != "https://api.launchpad.net/devel/~confirmer" {
		t.Errorf("Entries[1].OwnerLink = %q", got.Entries[1].OwnerLink)
	}
}

func TestMessageCollectionPagination(t *testing.T) {
	input := `{
		"total_size": 50,
		"start": 0,
		"next_collection_link": "https://api.launchpad.net/devel/bugs/1/messages?ws.start=25&ws.size=25",
		"entries": []
	}`

	var got MessageCollection
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.TotalSize != 50 {
		t.Errorf("TotalSize = %d, want 50", got.TotalSize)
	}
	if got.NextCollectionLink != "https://api.launchpad.net/devel/bugs/1/messages?ws.start=25&ws.size=25" {
		t.Errorf("NextCollectionLink = %q", got.NextCollectionLink)
	}
}
