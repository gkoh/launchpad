package launchpad

// CollectionMeta holds the pagination envelope common to all Launchpad
// API collection responses.
type CollectionMeta struct {
	TotalSize          int  `json:"total_size"`
	TotalSizeLink      Link `json:"total_size_link"`
	Start              int  `json:"start"`
	NextCollectionLink Link `json:"next_collection_link"`
	PrevCollectionLink Link `json:"prev_collection_link"`
}
