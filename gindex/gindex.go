package gindex

type SearchRequest struct {
	Token  string
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
