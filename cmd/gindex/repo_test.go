package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestRepoIndexing(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	var requests []http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Did receive the following:%v", r)
		requests = append(requests, *r)
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "commits") {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "blobs") {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
	}))
	defer ts.Close()
	fakeServer := ESServer{address: ts.URL}
	err := IndexRepoWithPath("../tdata/repo1.git", "tag1", &fakeServer, "testid", "testname")
	if err != nil {
		t.Errorf("Could  not open repository:%v", err)
	}
	log.Printf("Request n :%+v", requests)
}

func TestAnnexIndexing(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Did receive the following:%v", r)
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "commits") {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "blobs") {
			log.Printf("Need to reply with found")
			w.Write([]byte(`{"found": false}`))
		}

		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "blob/f78b7903bd67c78a98ccd4deffd3904dc0a3b431") {
			if r.ContentLength == 150 {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}))
	defer ts.Close()
	fakeServer := ESServer{address: ts.URL}
	err := IndexRepoWithPath("../tdata/repo2.git", "tag2", &fakeServer, "annextest", "testname")
	if err != nil {
		t.Errorf("Could  not open repository:%v", err)
	}
}
