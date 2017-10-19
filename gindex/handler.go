package gindex

import (
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"github.com/gogits/go-gogs-client"
	"io"
)

// Handler for Index requests
func IndexH(w http.ResponseWriter, r *http.Request, els *ElServer, rpath *string) {
	rbd := IndexRequest{}
	err := getParsedBody(r, &rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = IndexRepoWithPath(fmt.Sprintf("%s%s", *rpath, rbd.RepoPath),
		"master", els, rbd.RepoID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Handler for Search requests
func SearchH(w http.ResponseWriter, r *http.Request, els *ElServer, gins *GinServer) {
	rbd := SearchRequest{}
	err := getParsedBody(r, &rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get repo ids from the gin serevr to which the user has access
	// wer need tyo limit results to those
	repos := [] gogs.Repository{}
	err = getParsedResponse(http.MethodGet, fmt.Sprintf("%s/api/v1/user/repos", gins.URL),
		nil, rbd.Token, repos)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	repids := make([]string, len(repos))
	for c, repo := range repos {
		repids[c] = string(repo.ID)
	}
	w.WriteHeader(http.StatusOK)

}

func getParsedBody(r *http.Request, obj interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Debugf("Could not read request body: %+v", err)
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		log.Debugf("Could not unmarshal request: %+v, %s", err, string(data))
		return err
	}
	return nil
}

func getParsedResponse(method, path string, body io.Reader, token string, obj interface{}) error {
	client := &http.Client{}
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Cookie", fmt.Sprintf("i_like_gogits=%s", token))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf("Not Authorized")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}
