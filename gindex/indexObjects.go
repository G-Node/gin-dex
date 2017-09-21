package gindex

import (
	"encoding/json"
	"time"

	"io/ioutil"

	"log"

	"github.com/G-Node/gig"
)

type IndexBlob struct {
	*gig.Blob
	Repoid       string
	Id           int64
	GinRepoId    int64
	CommitSha    string
	Path         string
	Oid          int64
	IndexingTime time.Time
	Content      string
}

func NewCommitFromGig(gCommit *gig.Commit, repoid string) *IndexCommit {
	commit := &IndexCommit{gCommit, repoid, time.Now()}
	return commit
}

func NewBlobFromGig(gBlob *gig.Blob, repoid string) *IndexBlob {
	// Remember keeping the id
	blob := IndexBlob{Blob: gBlob, Repoid: repoid}
	return &blob
}

type IndexCommit struct {
	*gig.Commit
	GinRepoId    string
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

func (c *IndexCommit) AddToIndex(server *ElServer, index string) error {
	data, err := c.ToJson()
	if err != nil {
		return err
	}
	err = AddToIndex(data, server, index, "commit")
	return err
}

func (bl *IndexBlob) ToJson() ([]byte, error) {
	return json.Marshal(bl)
}

func (bl *IndexBlob) AddToIndex(server *ElServer, index string) error {
	f_type, err := DetermineFileType(bl)
	if err != nil {
		log.Printf("Could not determine file type:%+v", err)
		return nil
	}
	switch f_type {
	case TEXT:
		log.Printf("Text File found")
		ct, err := ioutil.ReadAll(bl)
		if err != nil {
			return err
		}
		bl.Content = string(ct)
		return nil
	case ODML_XML:
		ct, err := ioutil.ReadAll(bl)
		if err != nil {
			return err
		}
		bl.Content = string(ct)
	}
	data, err := bl.ToJson()
	if err != nil {
		return err
	}
	err = AddToIndex(data, server, index, "blob")
	return err
}

func (bl *IndexBlob) IsInIndex() (bool, error) {
	return false, nil
}

func AddToIndex(data []byte, server *ElServer, index, doctype string) error {
	_, err := server.Index(index, doctype, data)
	return err
}
