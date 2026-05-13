package launchpad

import "time"

// LockStatus describes the lock state of a bug.
type LockStatus string

const (
	LockStatusUnlocked    LockStatus = "Unlocked"
	LockStatusCommentOnly LockStatus = "Comment-only"
)

// Bug represents a bug entry from the Launchpad API.
type Bug struct {
	ActivityCollectionLink               Link            `json:"activity_collection_link"`
	AttachmentsCollectionLink            Link            `json:"attachments_collection_link"`
	BugTasksCollectionLink               Link            `json:"bug_tasks_collection_link"`
	BugWatchesCollectionLink             Link            `json:"bug_watches_collection_link"`
	CVEsCollectionLink                   Link            `json:"cves_collection_link"`
	DateCreated                          *time.Time      `json:"date_created,omitempty"`
	DateLastMessage                      *time.Time      `json:"date_last_message,omitempty"`
	DateLastUpdated                      *time.Time      `json:"date_last_updated,omitempty"`
	DateMadePrivate                      *time.Time      `json:"date_made_private,omitempty"`
	Description                          string          `json:"description"`
	DuplicateOfLink                      Link            `json:"duplicate_of_link"`
	DuplicatesCollectionLink             Link            `json:"duplicates_collection_link"`
	Heat                                 int             `json:"heat"`
	HTTPEtag                             string          `json:"http_etag"`
	ID                                   int             `json:"id"`
	InformationType                      InformationType `json:"information_type"`
	LatestPatchUploaded                  *time.Time      `json:"latest_patch_uploaded,omitempty"`
	LinkedBranchesCollectionLink         Link            `json:"linked_branches_collection_link"`
	LinkedMergeProposalsCollectionLink   Link            `json:"linked_merge_proposals_collection_link"`
	LockReason                           string          `json:"lock_reason,omitempty"`
	LockStatus                           LockStatus      `json:"lock_status"`
	MessageCount                         int             `json:"message_count"`
	MessagesCollectionLink               Link            `json:"messages_collection_link"`
	Name                                 string          `json:"name,omitempty"`
	NumberOfDuplicates                   int             `json:"number_of_duplicates"`
	OtherUsersAffectedCountWithDupes     int             `json:"other_users_affected_count_with_dupes"`
	OwnerLink                            Link            `json:"owner_link"`
	Private                              bool            `json:"private"`
	ResourceTypeLink                     Link            `json:"resource_type_link"`
	SecurityRelated                      bool            `json:"security_related"`
	SelfLink                             Link            `json:"self_link"`
	SubscriptionsCollectionLink          Link            `json:"subscriptions_collection_link"`
	Tags                                 []string        `json:"tags"`
	Title                                string          `json:"title"`
	UsersAffectedCollectionLink          Link            `json:"users_affected_collection_link"`
	UsersAffectedCount                   int             `json:"users_affected_count"`
	UsersAffectedCountWithDupes          int             `json:"users_affected_count_with_dupes"`
	UsersAffectedWithDupesCollectionLink Link            `json:"users_affected_with_dupes_collection_link"`
	UsersUnaffectedCollectionLink        Link            `json:"users_unaffected_collection_link"`
	UsersUnaffectedCount                 int             `json:"users_unaffected_count"`
	VulnerabilitiesCollectionLink        Link            `json:"vulnerabilities_collection_link"`
	WebLink                              Link            `json:"web_link"`
	WhoMadePrivateLink                   Link            `json:"who_made_private_link"`
}

// BugCollection is a paginated collection of Bug entries.
type BugCollection struct {
	CollectionMeta
	Entries []Bug `json:"entries"`
}

// BugTaskStatus describes the status of a bug task.
type BugTaskStatus string

const (
	BugTaskStatusNew          BugTaskStatus = "New"
	BugTaskStatusIncomplete   BugTaskStatus = "Incomplete"
	BugTaskStatusOpinion      BugTaskStatus = "Opinion"
	BugTaskStatusInvalid      BugTaskStatus = "Invalid"
	BugTaskStatusWontFix      BugTaskStatus = "Won't Fix"
	BugTaskStatusExpired      BugTaskStatus = "Expired"
	BugTaskStatusConfirmed    BugTaskStatus = "Confirmed"
	BugTaskStatusTriaged      BugTaskStatus = "Triaged"
	BugTaskStatusInProgress   BugTaskStatus = "In Progress"
	BugTaskStatusDeferred     BugTaskStatus = "Deferred"
	BugTaskStatusFixCommitted BugTaskStatus = "Fix Committed"
	BugTaskStatusFixReleased  BugTaskStatus = "Fix Released"
	BugTaskStatusDoesNotExist BugTaskStatus = "Does Not Exist"
	BugTaskStatusUnknown      BugTaskStatus = "Unknown"
)

// BugTaskImportance describes the importance of a bug task.
type BugTaskImportance string

const (
	BugTaskImportanceUnknown   BugTaskImportance = "Unknown"
	BugTaskImportanceUndecided BugTaskImportance = "Undecided"
	BugTaskImportanceCritical  BugTaskImportance = "Critical"
	BugTaskImportanceHigh      BugTaskImportance = "High"
	BugTaskImportanceMedium    BugTaskImportance = "Medium"
	BugTaskImportanceLow       BugTaskImportance = "Low"
	BugTaskImportanceWishlist  BugTaskImportance = "Wishlist"
)

// BugTask represents a bug task entry from the Launchpad API.
// A bug task tracks a bug needing fixing in a particular product or package.
type BugTask struct {
	AssigneeLink               Link              `json:"assignee_link"`
	BugLink                    Link              `json:"bug_link"`
	BugTargetDisplayName       string            `json:"bug_target_display_name"`
	BugTargetName              string            `json:"bug_target_name"`
	BugWatchLink               Link              `json:"bug_watch_link"`
	DateAssigned               *time.Time        `json:"date_assigned,omitempty"`
	DateClosed                 *time.Time        `json:"date_closed,omitempty"`
	DateConfirmed              *time.Time        `json:"date_confirmed,omitempty"`
	DateCreated                *time.Time        `json:"date_created,omitempty"`
	DateDeferred               *time.Time        `json:"date_deferred,omitempty"`
	DateFixCommitted           *time.Time        `json:"date_fix_committed,omitempty"`
	DateFixReleased            *time.Time        `json:"date_fix_released,omitempty"`
	DateInProgress             *time.Time        `json:"date_in_progress,omitempty"`
	DateIncomplete             *time.Time        `json:"date_incomplete,omitempty"`
	DateLeftClosed             *time.Time        `json:"date_left_closed,omitempty"`
	DateLeftNew                *time.Time        `json:"date_left_new,omitempty"`
	DateTriaged                *time.Time        `json:"date_triaged,omitempty"`
	HTTPEtag                   string            `json:"http_etag"`
	Importance                 BugTaskImportance `json:"importance"`
	ImportanceExplanation      string            `json:"importance_explanation,omitempty"`
	IsComplete                 bool              `json:"is_complete"`
	MilestoneLink              Link              `json:"milestone_link"`
	OwnerLink                  Link              `json:"owner_link"`
	RelatedTasksCollectionLink Link              `json:"related_tasks_collection_link"`
	ResourceTypeLink           Link              `json:"resource_type_link"`
	SelfLink                   Link              `json:"self_link"`
	Status                     BugTaskStatus     `json:"status"`
	StatusExplanation          string            `json:"status_explanation,omitempty"`
	TargetLink                 Link              `json:"target_link"`
	Title                      string            `json:"title"`
	WebLink                    Link              `json:"web_link"`
}

// BugTaskCollection is a paginated collection of BugTask entries.
type BugTaskCollection struct {
	CollectionMeta
	Entries []BugTask `json:"entries"`
}
