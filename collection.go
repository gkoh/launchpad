package launchpad

// CollectionMeta holds the pagination envelope common to all Launchpad
// API collection responses.
type CollectionMeta struct {
	TotalSize          int    `json:"total_size"`
	Start              int    `json:"start"`
	NextCollectionLink string `json:"next_collection_link,omitempty"`
	PrevCollectionLink string `json:"prev_collection_link,omitempty"`
}
