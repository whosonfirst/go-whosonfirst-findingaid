package repo

// ResolverResonse is a struct that contains Who's On First repository information for Who's On First records.
type FindingAidResponse struct {
	// The unique Who's On First ID.
	ID int64 `json:"id"`
	// The name of the Who's On First repository.
	Repo string `json:"repo"`
	// The relative path for a Who's On First ID.
	URI string `json:"uri"`
}
