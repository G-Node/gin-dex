package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/G-Node/gig"
	gannex "github.com/G-Node/go-annex"
	log "github.com/sirupsen/logrus"
)

// IndexBlob represents a git blob
type IndexBlob struct {
	*gig.Blob
	GinRepoName  string
	GinRepoId    string
	FirstCommit  string
	Id           int64
	Oid          gig.SHA1
	IndexingTime time.Time
	Content      string
	Path         string
}

func NewCommitFromGig(gCommit *gig.Commit, repoid string, reponame string, oid gig.SHA1) *IndexCommit {
	commit := &IndexCommit{gCommit, repoid, oid,
		reponame, time.Now()}
	return commit
}

func NewBlobFromGig(gBlob *gig.Blob, repoid string, oid gig.SHA1, commit string, path string, reponame string) *IndexBlob {
	// Remember keeping the id
	blob := IndexBlob{Blob: gBlob, GinRepoId: repoid, Oid: oid, FirstCommit: commit, Path: path, GinRepoName: reponame}
	return &blob
}

type IndexCommit struct {
	*gig.Commit
	GinRepoId    string
	Oid          gig.SHA1
	GinRepoName  string
	IndexingTime time.Time
}

func BlobFromJson(data []byte) (*IndexBlob, error) {
	bl := &IndexBlob{}
	err := json.Unmarshal(data, bl)
	return bl, err
}

func (c *IndexCommit) ToJson() ([]byte, error) {
	return json.Marshal(c)
}

func (c *IndexCommit) AddToIndex(essrv *ESServer, index string, id gig.SHA1) error {
	data, err := c.ToJson()
	if err != nil {
		return err
	}
	indexid := GetIndexCommitId(id.String(), c.GinRepoId)
	err = AddToIndex(data, essrv, index, "commit", indexid)
	return err
}

func (bl *IndexBlob) ToJson() ([]byte, error) {
	return json.Marshal(bl)
}

func (bl *IndexBlob) AddToIndexTimeout(cfg *Configuration, id gig.SHA1) error {
	err := make(chan error)
	defer close(err)
	go func() { err <- bl.AddToIndex(cfg, id) }()
	select {
	case res := <-err:
		return res
	case <-time.After(time.Duration(cfg.Timeout) * time.Second):
		return fmt.Errorf("Timed out: %s, %v", cfg.RepositoryStore, bl)
	}
}

func (bl *IndexBlob) AddToIndex(cfg *Configuration, id gig.SHA1) error {
	indexid := GetIndexCommitId(id.String(), bl.GinRepoId)
	ftype, blobBuffer, err := BlobFileType(bl)
	if err != nil {
		log.Errorf("Could not determine file type: %v", err)
		return nil
	}

	switch ftype {
	case ANNEX:
		APFileC, err := ioutil.ReadAll(blobBuffer)
		log.Debugf("Annex file: %s", APFileC)
		if err != nil {
			log.Errorf("Could not open annex pointer file: %v", err)
			return err
		}
		annexfile, err := gannex.NewAFile(cfg.RepositoryStore, "", "", APFileC)
		if err != nil {
			log.Errorf("Could not get annex file: %v", err)
			return err
		}
		fp, err := annexfile.Open()
		if err != nil {
			log.Errorf("Could not open annex file: %v", err)
			return err
		}
		defer fp.Close()
		bl.Blob = gig.MakeAnnexBlob(fp, annexfile.Info.Size())
		return bl.AddToIndex(cfg, id)

	case TEXT:
		if bl.Size() > gannex.MEGABYTE*cfg.MaxTextSize {
			return fmt.Errorf("File to big")
		}
		ct, err := ioutil.ReadAll(blobBuffer)
		if err != nil {
			log.Errorf("Could not read text file content: %v", err)
			return err
		}
		bl.Content = string(ct)
	case ODML_XML:
		ct, err := ioutil.ReadAll(blobBuffer)
		if err != nil {
			return err
		}
		bl.Content = string(ct)
	case PDF:
		if bl.Size() > gannex.MEGABYTE*cfg.MaxPDFSize {
			return fmt.Errorf("File to big")
		}
		content, err := GetPlainPdf(blobBuffer, bl.Size())
		if err != nil {
			log.Debugf("Could not read pdf: %v", err)
			return err
		}
		bl.Content = content
	case NEV:
		// Get the main nev comemnts
		com, err := GetNevComments(blobBuffer)
		if err != nil {
			return err
		}
		bl.Content = *com
	}

	data, err := bl.ToJson()
	if err != nil {
		return err
	}
	err = AddToIndex(data, cfg.Elasticsearch, cfg.Elasticsearch.blindex, "blob", indexid)
	return err
}

func (bl *IndexBlob) IsInIndex() (bool, error) {
	return false, nil
}

func AddToIndex(data []byte, server *ESServer, index, doctype string, id gig.SHA1) error {
	resp, err := server.Index(index, doctype, data, id)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return err
}
