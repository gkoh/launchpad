package launchpad

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestAccountStatusJSON(t *testing.T) {
	tests := []struct {
		val  AccountStatus
		want string
	}{
		{AccountStatusPlaceholder, `"Placeholder"`},
		{AccountStatusUnactivated, `"Unactivated"`},
		{AccountStatusActive, `"Active"`},
		{AccountStatusDeactivated, `"Deactivated"`},
		{AccountStatusSuspended, `"Suspended"`},
		{AccountStatusClosed, `"Closed"`},
		{AccountStatusDeceased, `"Deceased"`},
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

		var got AccountStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestPersonJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	orig := Person{
		AccountStatus:    AccountStatusActive,
		DateCreated:      &now,
		Description:      "A test user",
		DisplayName:      "Test User",
		HTTPEtag:         "\"etag789\"",
		IsTeam:           false,
		IsValid:          true,
		Karma:            42,
		Name:             "test-user",
		ResourceTypeLink: "https://api.launchpad.net/devel/#person",
		SelfLink:         "https://api.launchpad.net/devel/~test-user",
		WebLink:          "https://launchpad.net/~test-user",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Person
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", got, orig)
	}
}

func TestPersonJSONNulls(t *testing.T) {
	input := `{
		"account_status": "Active",
		"display_name": "Minimal User",
		"http_etag": "",
		"is_team": false,
		"is_valid": true,
		"karma": 0,
		"name": "minimal",
		"resource_type_link": "",
		"self_link": "",
		"web_link": ""
	}`

	var got Person
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.DisplayName != "Minimal User" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "Minimal User")
	}
	if got.DateCreated != nil {
		t.Errorf("DateCreated = %v, want nil", got.DateCreated)
	}
	if got.Description != "" {
		t.Errorf("Description = %q, want empty", got.Description)
	}
	if got.AccountStatus != AccountStatusActive {
		t.Errorf("AccountStatus = %q, want %q", got.AccountStatus, AccountStatusActive)
	}
}
