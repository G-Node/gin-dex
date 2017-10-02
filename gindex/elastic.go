package gindex

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"

	"io/ioutil"

	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

type ElServer struct {
	adress   string
	uname    *string
	password *string
}

func (el *ElServer) Index(index, doctype string, data []byte, id gig.SHA1) (*http.Response, error) {
	adrr := fmt.Sprintf("%s/%s/%s/%s", el.adress, index, doctype, id.String())
	req, err := http.NewRequest("POST", adrr, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ElServer) elasticRequest(req *http.Request) (*http.Response, error) {
	if el.uname != nil {
		req.SetBasicAuth(*el.uname, *el.password)
	}
	req.Header.Set("Content-Type", "application/json")
	cl := http.Client{}
	log.Debugf("Indexing request:%+v", req)
	return cl.Do(req)
}

func (el *ElServer) HasCommit(index string, commitId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/commits/commit/%s", el.adress, commitId)
	return el.Has(adrr)
}

func (el *ElServer) HasBlob(index string, blobId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/blobs/blob/%s", el.adress, blobId)
	return el.Has(adrr)
}

func (el *ElServer) Has(adr string) (bool, error) {
	req, err := http.NewRequest("GET", adr, nil)
	if err != nil {
		return false, err
	}
	resp, err := el.elasticRequest(req)
	if err != nil {
		return false, err
	}
	bdy, err := ioutil.ReadAll(resp.Body)
	var res struct{ Found bool }
	err = json.Unmarshal(bdy, &res)
	if err != nil {
		log.WithError(err)
		return false, err
	}
	return res.Found, nil
}
