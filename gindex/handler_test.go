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
	log.SetLevel(log.ErrorLevel)
	rbd := IndexRequest{Token: "testtoken", UserID: 10, RepoPath: "repo2.git",
		RepoID: "repo2.git"}
	data, err := json.Marshal(rbd)
	if err != nil {
		log.Debugf("could not marshal Index request:%+v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("did receive the following:%v", r)
		if (r.Method == http.MethodGet && strings.Contains(r.URL.Path, "commits")) {
			log.Printf("need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
		if (r.Method == http.MethodGet && strings.Contains(r.URL.Path, "blobs")) {
			log.Printf("need to reply with found")
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

func TestSearchHandler(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	rbd := SearchRequest{Token: "testtoken", UserID: 10, Querry: "Test Search"}
	data, err := json.Marshal(rbd)
	if err != nil {
		log.Debugf("could not marshal search request:%+v", err)
	}

	// Test with result and with error
	searchResultBlob := `{"took":3,"timed_out":false,"_shards":{"total":5,"successful":5,"skipped":0,"failed":0},"hits":{"total":3,"max_score":0.2824934,"hits":[{"_index":"blobs","_type":"blob","_id":"123456","_score":0.2824934,"_source":{"Repoid":"repo2.git","Id":0,"GinRepoId":0,"CommitSha":"","Path":"","Oid":0,"IndexingTime":"0001-01-01T00:00:00Z","Content":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diamn"}},{"_index":"blobs","_type":"blob","_id":"123457","_score":0.2824934,"_source":{"Repoid":"repo2.git","Id":0,"GinRepoId":0,"CommitSha":"","Path":"","Oid":0,"IndexingTime":"0001-01-01T00:00:00Z","Content":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diamn"}},{"_index":"blobs","_type":"blob","_id":"12345","_score":0.2824934,"_source":{"Repoid":"repo2.git","Id":0,"GinRepoId":0,"CommitSha":"","Path":"","Oid":0,"IndexingTime":"0001-01-01T00:00:00Z","Content":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diamn"}}]}}`
	searchResultCommit := `{"error":{"root_cause":[{"type":"index_not_found_exception","reason":"no such index","resource.type":"index_or_alias","resource.id":"index","index_uuid":"_na_","index":"index"}],"type":"index_not_found_exception","reason":"no such index","resource.type":"index_or_alias","resource.id":"index","index_uuid":"_na_","index":"index"},"status":404}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	ts := makeFakeServer(searchResultBlob, searchResultCommit)
	rec := httptest.NewRecorder()
	SearchH(rec, req, &ElServer{adress: ts.URL}, &GinServer{URL: ts.URL})
	if rec.Code != http.StatusOK {
		t.Fail()
		return
	}

	// Test with result and with error and empty result
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	searchResultBlob = `{"took":1,"timed_out":false,"_shards":{"total":5,"successful":5,"skipped":0,"failed":0},"hits":{"total":0,"max_score":null,"hits":[]}}`
	ts = makeFakeServer(searchResultBlob, searchResultCommit)
	rec = httptest.NewRecorder()
	SearchH(rec, req, &ElServer{adress: ts.URL}, &GinServer{URL: ts.URL})
	if rec.Code != http.StatusOK {
		t.Fail()
		return
	}
	t.Logf("[OK] searching")
}

func makeFakeServer(searchResultBlob, searchResultCommit string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("did receive the following:%v", r)
		if (r.Method == http.MethodPost && strings.Contains(r.URL.Path, "commits")) {
			w.Write([]byte(searchResultCommit))
		}
		if (r.Method == http.MethodPost && strings.Contains(r.URL.Path, "blobs")) {
			log.Printf("need to reply with blob results")
			w.Write([]byte(searchResultBlob))
		}
		if (r.Method == http.MethodGet && strings.Contains(r.URL.Path, "api/v1/user/repos")) {
			w.Write([]byte(`[{"id":0}]`))
		}
	}))
}