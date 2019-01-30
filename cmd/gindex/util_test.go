package gindex

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/G-Node/gig"
	log "github.com/sirupsen/logrus"
)

func TestHasRepoAccess(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	rbd := ReIndexRequest{}
	gRepo := gig.Repository{Path: "/home/test/bla/franz/repo1"}
	ts := makeFakeUtilServer()
	repo, err := hasRepoAccess(&gRepo, &rbd, &GinServer{URL: ts.URL})
	if err != nil {
		t.Logf("%+v", err)
		t.Fail()
		return
	}
	log.Debugf("repso is:%+v", repo)
	t.Logf("[OK] repo acces")
}

func TestFindRepos(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	rbd := ReIndexRequest{Token: "", CsrfT: ""}
	ts := makeFakeUtilServer()
	repos, err := findRepos("../tdata", &rbd, &GinServer{URL: ts.URL})
	if err != nil {
		t.Logf("%+v", err)
		t.Fail()
		return
	}
	log.Debugf("repos are:%+v", repos)
	t.Logf("[OK] repo finding")
}

func makeFakeUtilServer() *httptest.Server {
	repoReturn := `{"id":44,"owner":{"id":1,"login":"testi","full_name":"tetsi","email":"ss@gn.com","avatar_url":"https://gin.g-node.org/avatars/1","username":"cgars"},"name":"annex","full_name":"test/test","description":"","private":true,"fork":false,"parent":null,"empty":false,"mirror":false,"size":3239936,"html_url":"","ssh_url":"","clone_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"default_branch":"master","created_at":"2017-06-20T08:12:56Z","updated_at":"2017-06-23T15:48:23Z","permissions":{"admin":true,"push":true,"pull":true}}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("did receive the following:%v", r)
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "api/v1/repos") {
			w.Write([]byte(repoReturn))
		}
	}))
}
