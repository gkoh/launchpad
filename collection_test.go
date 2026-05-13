package launchpad

import (
	"encoding/json"
	"testing"
)

func TestCollectionMetaJSON(t *testing.T) {
	orig := CollectionMeta{
		TotalSize:          42,
		Start:              10,
		NextCollectionLink: NewLink("https://api.launchpad.net/devel/branches?ws.start=20"),
		PrevCollectionLink: NewLink("https://api.launchpad.net/devel/branches?ws.start=0"),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got CollectionMeta
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if orig.TotalSize != got.TotalSize {
		t.Errorf("TotalSize = %d, want %d", got.TotalSize, orig.TotalSize)
	}
	if orig.Start != got.Start {
		t.Errorf("Start = %d, want %d", got.Start, orig.Start)
	}
	if orig.NextCollectionLink.String() != got.NextCollectionLink.String() {
		t.Errorf("NextCollectionLink = %q, want %q", got.NextCollectionLink, orig.NextCollectionLink)
	}
	if orig.PrevCollectionLink.String() != got.PrevCollectionLink.String() {
		t.Errorf("PrevCollectionLink = %q, want %q", got.PrevCollectionLink, orig.PrevCollectionLink)
	}
}

func TestCollectionMetaJSONZeroLinks(t *testing.T) {
	orig := CollectionMeta{
		TotalSize: 5,
		Start:     0,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got CollectionMeta
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !got.NextCollectionLink.IsZero() {
		t.Errorf("NextCollectionLink = %q, want zero", got.NextCollectionLink)
	}
	if !got.PrevCollectionLink.IsZero() {
		t.Errorf("PrevCollectionLink = %q, want zero", got.PrevCollectionLink)
	}
}

// jsonContainsKey checks whether a JSON string contains a given key.
func jsonContainsKey(raw, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
