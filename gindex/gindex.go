package gindex

const (
	SEARCH_MATCH = 0
	SEARCH_FUZZY = 1
	SEARCH_WILDCARD = 2
)
type SearchRequest struct {
	Token  string
	CsrfT  string
	UserID int64
	Querry string
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
	URL     string
	GetRepo string
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

var Setting struct {
	TxtMSize int64
	PdfMSize int64
	Timeout  int64
}
