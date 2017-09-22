package gindex

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoIndexing(t *testing.T) {
	var requests []http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Did receive the following:%v", r)
		ct, _ := ioutil.ReadAll(r.Body)
		log.Printf("Did receive the following content:%v", string(ct))
		requests = append(requests, *r)
	}))
	defer ts.Close()
	fakeServer := ElServer{ts.URL}
	err := IndexRepoWithPath("../tdata/repo1.git", &fakeServer, "testid")
	if err != nil {
		t.Errorf("Could  not open repository:%v", err)
	}
	log.Printf("Request n :%+v", requests)
}
