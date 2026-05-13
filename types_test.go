package launchpad

import (
	"encoding/json"
	"testing"
)

func TestInformationTypeJSON(t *testing.T) {
	tests := []struct {
		val  InformationType
		want string
	}{
		{InformationPublic, `"Public"`},
		{InformationPublicSecurity, `"Public Security"`},
		{InformationPrivateSecurity, `"Private Security"`},
		{InformationPrivate, `"Private"`},
		{InformationProprietary, `"Proprietary"`},
		{InformationEmbargoed, `"Embargoed"`},
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

		var got InformationType
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", data, err)
			continue
		}
		if got != tt.val {
			t.Errorf("Unmarshal(%s) = %q, want %q", data, got, tt.val)
		}
	}
}

func TestInformationTypeUnmarshalUnknown(t *testing.T) {
	var got InformationType
	if err := json.Unmarshal([]byte(`"Something New"`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != "Something New" {
		t.Errorf("got %q, want %q", got, "Something New")
	}
}
