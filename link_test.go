package launchpad

import (
	"encoding/json"
	"testing"
)

func TestLinkMarshalJSON(t *testing.T) {
	l := NewLink("https://api.launchpad.net/devel/bugs/1")
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"https://api.launchpad.net/devel/bugs/1"` {
		t.Errorf("got %s", data)
	}
}

func TestLinkMarshalJSONZero(t *testing.T) {
	var l Link
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `""` {
		t.Errorf("got %s, want empty string", data)
	}
}

func TestLinkUnmarshalJSON(t *testing.T) {
	var l Link
	if err := json.Unmarshal([]byte(`"https://api.launchpad.net/devel/bugs/1"`), &l); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if l.IsZero() {
		t.Fatal("expected non-zero link")
	}
	if l.String() != "https://api.launchpad.net/devel/bugs/1" {
		t.Errorf("got %q", l.String())
	}
}

func TestLinkUnmarshalJSONNull(t *testing.T) {
	var l Link
	if err := json.Unmarshal([]byte(`null`), &l); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !l.IsZero() {
		t.Error("expected zero link")
	}
}

func TestLinkUnmarshalJSONEmpty(t *testing.T) {
	var l Link
	if err := json.Unmarshal([]byte(`""`), &l); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !l.IsZero() {
		t.Error("expected zero link")
	}
}

func TestLinkIsZero(t *testing.T) {
	var zero Link
	if !zero.IsZero() {
		t.Error("zero value should be zero")
	}
	l := NewLink("https://example.com")
	if l.IsZero() {
		t.Error("non-empty link should not be zero")
	}
}

func TestLinkString(t *testing.T) {
	var zero Link
	if zero.String() != "" {
		t.Errorf("zero String() = %q, want empty", zero.String())
	}
	l := NewLink("https://example.com/path")
	if l.String() != "https://example.com/path" {
		t.Errorf("String() = %q", l.String())
	}
}

func TestLinkRoundTrip(t *testing.T) {
	orig := NewLink("https://api.launchpad.net/devel/~user")
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Link
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if orig.String() != got.String() {
		t.Errorf("round-trip mismatch: got %q, want %q", got.String(), orig.String())
	}
}

func TestNewLink(t *testing.T) {
	l := NewLink("")
	if !l.IsZero() {
		t.Error("NewLink(\"\") should be zero")
	}
}

func TestNewLinkPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid URL")
		}
	}()
	NewLink("://invalid")
}
