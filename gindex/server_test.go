	package gindex

import (
	"testing"

	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

func TestServerIndexing(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	pw := "changeme"
	un := "elastic"
	fakeServer := ElServer{adress: "http://127.0.0.1:9200", uname: &un, password: &pw}
	err := IndexRepoWithPath("../tdata/repo1.git", "tag1", &fakeServer, "testid")
	if err != nil {
		t.Errorf("Could  not index repo:%v", err)
	}

}

func TestServerSearch(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	pw := "changeme"
	un := "elastic"
	fakeServer := ElServer{adress: "http://127.0.0.1:9200", uname: &un, password: &pw}
	res, err := fakeServer.SearchBlobs("christian", "commits", []string{"testid", "hhjh"})
	if err != nil {
		t.Errorf("Could  not index repo:%v", err)
	}
	bd, _ := ioutil.ReadAll(res.Body)
	log.Printf("Result is:%+v", res)
	log.Printf("Body is:%+v", string(bd))
}
