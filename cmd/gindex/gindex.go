package main

const (
	SEARCH_MATCH = iota
	SEARCH_FUZZY
	SEARCH_WILDCARD
	SEARCH_QUERY
	SEARCH_SUGGEST
)

type SearchRequest struct {
	Token  string
	CsrfT  string
	UserID int64
	Query  string
	SType  int64
}

type IndexRequest struct {
	UserID   int
	RepoPath string
	RepoID   string
}

type ReIndexRequest struct {
	*IndexRequest
	Token string
	CsrfT string
}
type GinServer struct {
	URL string
	Key string
}

type BlobSResult struct {
	Source    *IndexBlob  `json:"_source"`
	Score     float64     `json:"_score"`
	Highlight interface{} `json:"highlight"`
}

type CommitSResult struct {
	Source    *IndexCommit `json:"_source"`
	Score     float64      `json:"_score"`
	Highlight interface{}  `json:"highlight"`
}

type SearchResults struct {
	Blobs   []BlobSResult
	Commits []CommitSResult
}

type Suggestions struct {
	Items []Suggestion
}

type Suggestion struct {
	Title string
	Url   string
}
