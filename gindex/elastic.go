package gindex

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/G-Node/gig"
)

type ElServer struct {
	adress   string
	uname    *string
	password *string
}

func (el *ElServer) Index(index, doctype string, data []byte) (*http.Response, error) {
	adrr := fmt.Sprintf("%s/%s/%s", el.adress, index, doctype)
	req, err := http.NewRequest("PUT", adrr, bytes.NewReader(data))
	if el.uname != nil {
		req.SetBasicAuth(*el.uname, *el.password)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	cl := http.Client{}
	return cl.Do(req)

}

func (el *ElServer) HasCommit(index string, commitId gig.SHA1) (bool, error) {
	return false, nil
}

func (el *ElServer) HasBlob(index string, blobId gig.SHA1) (bool, error) {
	return false, nil
}
