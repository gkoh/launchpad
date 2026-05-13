package launchpad

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCollectionMetaJSON(t *testing.T) {
	orig := CollectionMeta{
		TotalSize:          42,
		Start:              10,
		NextCollectionLink: "https://api.launchpad.net/devel/branches?ws.start=20",
		PrevCollectionLink: "https://api.launchpad.net/devel/branches?ws.start=0",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got CollectionMeta
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", got, orig)
	}
}

func TestCollectionMetaJSONOmitEmpty(t *testing.T) {
	orig := CollectionMeta{
		TotalSize: 5,
		Start:     0,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	raw := string(data)
	for _, key := range []string{"next_collection_link", "prev_collection_link"} {
		if contains := jsonContainsKey(raw, key); contains {
			t.Errorf("expected key %q to be omitted, got JSON: %s", key, raw)
		}
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
