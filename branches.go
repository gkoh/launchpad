package launchpad

import "time"

// BranchType describes how a Bazaar branch is hosted.
type BranchType string

const (
	BranchTypeHosted   BranchType = "Hosted"
	BranchTypeMirrored BranchType = "Mirrored"
	BranchTypeImported BranchType = "Imported"
	BranchTypeRemote   BranchType = "Remote"
)

// LifecycleStatus describes the development status of a branch.
type LifecycleStatus string

const (
	LifecycleExperimental LifecycleStatus = "Experimental"
	LifecycleDevelopment  LifecycleStatus = "Development"
	LifecycleMature       LifecycleStatus = "Mature"
	LifecycleMerged       LifecycleStatus = "Merged"
	LifecycleAbandoned    LifecycleStatus = "Abandoned"
)

// Branch represents a Bazaar branch entry from the Launchpad API.
type Branch struct {
	BranchFormat                    string          `json:"branch_format"`
	BranchType                      BranchType      `json:"branch_type"`
	BzrIdentity                     string          `json:"bzr_identity"`
	CodeImportLink                  Link            `json:"code_import_link"`
	ControlFormat                   string          `json:"control_format"`
	DateCreated                     *time.Time      `json:"date_created,omitempty"`
	DateLastModified                *time.Time      `json:"date_last_modified,omitempty"`
	DependentBranchesCollectionLink Link            `json:"dependent_branches_collection_link"`
	Description                     string          `json:"description"`
	DisplayName                     string          `json:"display_name"`
	ExplicitlyPrivate               bool            `json:"explicitly_private"`
	HTTPEtag                        string          `json:"http_etag"`
	InformationType                 InformationType `json:"information_type"`
	LandingCandidatesCollectionLink Link            `json:"landing_candidates_collection_link"`
	LandingTargetsCollectionLink    Link            `json:"landing_targets_collection_link"`
	LastMirrorAttempt               *time.Time      `json:"last_mirror_attempt,omitempty"`
	LastMirrored                    *time.Time      `json:"last_mirrored,omitempty"`
	LastScanned                     *time.Time      `json:"last_scanned,omitempty"`
	LastScannedID                   string          `json:"last_scanned_id"`
	LifecycleStatus                 LifecycleStatus `json:"lifecycle_status"`
	LinkedBugsCollectionLink        Link            `json:"linked_bugs_collection_link"`
	MirrorStatusMessage             string          `json:"mirror_status_message"`
	Name                            string          `json:"name"`
	OwnerLink                       Link            `json:"owner_link"`
	Private                         bool            `json:"private"`
	ProjectLink                     Link            `json:"project_link"`
	RecipesCollectionLink           Link            `json:"recipes_collection_link"`
	RegistrantLink                  Link            `json:"registrant_link"`
	RepositoryFormat                string          `json:"repository_format"`
	ResourceTypeLink                Link            `json:"resource_type_link"`
	ReviewerLink                    Link            `json:"reviewer_link"`
	RevisionCount                   int             `json:"revision_count"`
	SelfLink                        Link            `json:"self_link"`
	SourcePackageLink               Link            `json:"sourcepackage_link"`
	SpecLinksCollectionLink         Link            `json:"spec_links_collection_link"`
	SubscribersCollectionLink       Link            `json:"subscribers_collection_link"`
	SubscriptionsCollectionLink     Link            `json:"subscriptions_collection_link"`
	UniqueName                      string          `json:"unique_name"`
	URL                             Link            `json:"url"`
	WebLink                         Link            `json:"web_link"`
	WebhooksCollectionLink          Link            `json:"webhooks_collection_link"`
	Whiteboard                      string          `json:"whiteboard,omitempty"`
}

// BranchCollection is a paginated collection of Branch entries.
type BranchCollection struct {
	CollectionMeta
	Entries []Branch `json:"entries"`
}
