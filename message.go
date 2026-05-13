package launchpad

import "time"

// Message represents a message (comment) on a Launchpad bug.
type Message struct {
	Content          string     `json:"content"`
	DateCreated      *time.Time `json:"date_created,omitempty"`
	HTTPEtag         string     `json:"http_etag"`
	OwnerLink        Link       `json:"owner_link"`
	ResourceTypeLink Link       `json:"resource_type_link"`
	SelfLink         Link       `json:"self_link"`
	Subject          string     `json:"subject"`
	WebLink          Link       `json:"web_link"`
}

// MessageCollection is a paginated collection of Message entries.
type MessageCollection struct {
	CollectionMeta
	Entries []Message `json:"entries"`
}
