package launchpad

import "time"

// Message represents a message (comment) on a Launchpad bug.
type Message struct {
	Content          string     `json:"content"`
	DateCreated      *time.Time `json:"date_created,omitempty"`
	HTTPEtag         string     `json:"http_etag"`
	OwnerLink        string     `json:"owner_link"`
	ResourceTypeLink string     `json:"resource_type_link"`
	SelfLink         string     `json:"self_link"`
	Subject          string     `json:"subject"`
	WebLink          string     `json:"web_link"`
}

// MessageCollection is a paginated collection of Message entries.
type MessageCollection struct {
	CollectionMeta
	Entries []Message `json:"entries"`
}
