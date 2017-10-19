package gindex

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"encoding/json"
	"bytes"
	log "github.com/Sirupsen/logrus"
	"strings"
)

func TestIndexHandler(t *testing.T) {
	rbd := IndexRequest{Token: "testtoken", UserID: 10, RepoPath: "repo2.git",
		RepoID: "repo2.git"}
	data, err := json.Marshal(rbd)
	if err != nil {
		log.Debugf("Could not marshal Index request:%+v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Did receive the following:%v", r)
		if (r.Method == http.MethodGet && strings.Contains(r.URL.Path, "commits")) {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
		if (r.Method == http.MethodGet && strings.Contains(r.URL.Path, "blobs")) {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
	}))
	el := ElServer{adress: ts.URL}
	repopath := "../tdata/"
	IndexH(rec, req, &el, &repopath)
	log.Debugf("IndexH Response: %+v", rec)
	// todo: test more than just Status gfn
	if rec.Code != http.StatusOK {
		t.Fail()
	}
	t.Log("[OK] indexing")
}
