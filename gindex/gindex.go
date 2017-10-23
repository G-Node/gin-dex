package gindex

type SearchRequest struct {
	Token  string
	CsrfT  string
	UserID int
	Querry string
}

type IndexRequest struct {
	Token    string
	UserID   int
	RepoPath string
	RepoID   string
}
type GinServer struct {
	URL     string
	GetRepo string
}

type BlobSResult struct {
	Source *IndexBlob `json:"_source"`
	Score  float64    `json:"_score"`
}

type CommitSResult struct {
	Source *IndexCommit `json:"_source"`
	Score  float64    `json:"_score"`
}

type SearchResults struct {
	Blobs []BlobSResult
	Commits []CommitSResult
}