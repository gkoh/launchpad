package launchpad

import "time"

// AccountStatus describes the status of a person's Launchpad account.
type AccountStatus string

const (
	AccountStatusPlaceholder AccountStatus = "Placeholder"
	AccountStatusUnactivated AccountStatus = "Unactivated"
	AccountStatusActive      AccountStatus = "Active"
	AccountStatusDeactivated AccountStatus = "Deactivated"
	AccountStatusSuspended   AccountStatus = "Suspended"
	AccountStatusClosed      AccountStatus = "Closed"
	AccountStatusDeceased    AccountStatus = "Deceased"
)

// Person represents a person or team entry from the Launchpad API.
type Person struct {
	AccountStatus    AccountStatus `json:"account_status"`
	DateCreated      *time.Time    `json:"date_created,omitempty"`
	Description      string        `json:"description,omitempty"`
	DisplayName      string        `json:"display_name"`
	HTTPEtag         string        `json:"http_etag"`
	IsTeam           bool          `json:"is_team"`
	IsValid          bool          `json:"is_valid"`
	Karma            int           `json:"karma"`
	Name             string        `json:"name"`
	ResourceTypeLink Link          `json:"resource_type_link"`
	SelfLink         Link          `json:"self_link"`
	WebLink          Link          `json:"web_link"`
}
