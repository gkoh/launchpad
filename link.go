package launchpad

import (
	"encoding/json"
	"net/url"
)

// Link is a URL reference from the Launchpad API. It handles JSON
// marshalling between string URLs and *url.URL values. A zero-value
// Link represents an absent or null link.
type Link struct {
	*url.URL
}

// NewLink parses s into a Link. It panics if s is not a valid URL,
// so it should only be used with string literals or trusted input.
func NewLink(s string) Link {
	if s == "" {
		return Link{}
	}
	u, err := url.Parse(s)
	if err != nil {
		panic("launchpad.NewLink: " + err.Error())
	}
	return Link{u}
}

// IsZero returns true if the link is nil/empty.
func (l Link) IsZero() bool {
	return l.URL == nil
}

// String returns the URL string, or "" if nil.
func (l Link) String() string {
	if l.URL == nil {
		return ""
	}
	return l.URL.String()
}

// MarshalJSON encodes the link as a JSON string, or null if empty.
func (l Link) MarshalJSON() ([]byte, error) {
	if l.URL == nil {
		return []byte(`""`), nil
	}
	return json.Marshal(l.URL.String())
}

// UnmarshalJSON decodes a JSON string into the link.
func (l *Link) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil || *s == "" {
		l.URL = nil
		return nil
	}
	u, err := url.Parse(*s)
	if err != nil {
		return err
	}
	l.URL = u
	return nil
}
