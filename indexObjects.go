package gindex

import (
	"github.com/G-Node/gin-dex/git"
	"time"
	"encoding/json"
)

type IndexBlob struct {
	git.Blob
	Id           int64
	GinRepoId    int64
	CommitSha    string
	Path         string
	Oid          int64
	Content      string
	IndexingTime time.Time
}

type IndexCommit struct {
	git.Commit
	GinRepoId    int64
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
