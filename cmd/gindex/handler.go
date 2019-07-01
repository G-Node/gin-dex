package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"

	"github.com/G-Node/libgin/libgin"
	log "github.com/sirupsen/logrus"
)

// Handler for Index requests
func indexHandler(w http.ResponseWriter, r *http.Request, cfg *Configuration) {
	rpath := cfg.RepositoryStore
	rbd := IndexRequest{}
	err := getParsedBody(r, cfg.Key, &rbd)
	log.Debugf("Got an indexing request: %+v", rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	repo := strings.ToLower(rbd.RepoPath)
	if repo[len(repo)-4:] != ".git" {
		repo = repo + ".git"
	}
	err = IndexRepoWithPath(cfg, fmt.Sprintf("%s/%s", rpath, repo), "master", rbd.RepoID, rbd.RepoPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// Handler for SearchBlobs requests
func searchHandler(w http.ResponseWriter, r *http.Request, cfg *Configuration) {
	els := cfg.Elasticsearch
	sreq := &libgin.SearchRequest{}
	err := getParsedBody(r, cfg.Key, &sreq)
	if err != nil {
		log.Errorf("Could not read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("Repos to search in [search]: %+v", sreq.RepoIDs)
	if sreq.SType == SEARCH_SUGGEST {
		suggestions, err := suggest(sreq, els)
		if err != nil {
			log.Warnf("Could not search blobs: %v", err)
		}
		result := []Suggestion{}
		for _, suf := range suggestions {
			result = append(result, Suggestion{Title: suf})
		}
		suggestionsJ, err := json.Marshal(Suggestions{Items: result})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(suggestionsJ)
		return
	}
	// Lets search now
	rBlobs := []BlobSResult{}
	log.Debug("Searching blobs")
	err = searchBlobs(sreq, els, &rBlobs)
	if err != nil {
		log.Warnf("Could not search blobs: %v", err)
	}
	rCommits := []CommitSResult{}
	log.Debug("Searching commits")
	err = searchCommits(sreq, els, &rCommits)
	if err != nil {
		log.Warnf("Could not search commits: %v", err)
	}

	data, err := encodeResponse(&SearchResults{Blobs: rBlobs, Commits: rCommits}, cfg.Key)
	if err != nil {
		log.Debugf("Could not marshal search results")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("[Matches] Blobs: %d  Commits: %d", len(rBlobs), len(rCommits))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func suggestHandler(w http.ResponseWriter, r *http.Request, els *ESServer) {
	sreq := &libgin.SearchRequest{}
	log.Debugf("Repos to search in [suggest]: %+v", sreq.RepoIDs)
	// Lets search now
	suggestions, err := suggest(sreq, els)
	if err != nil {
		log.Warnf("Could not search blobs: %v", err)
	}
	suggestionsJ, err := json.Marshal(suggestions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(suggestionsJ)
}

// Handler for Index requests
func reIndexRepo(w http.ResponseWriter, r *http.Request, cfg *Configuration) {
	rbd := IndexRequest{}
	err := getParsedBody(r, cfg.Key, &rbd)
	log.Debugf("Got an indexing request: %+v", rbd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = reIndexRepoWithPath(cfg, fmt.Sprintf("%s/%s", cfg.RepositoryStore, strings.ToLower(rbd.RepoPath)+".git"), "master", rbd.RepoID, rbd.RepoPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
func reIndexHandler(w http.ResponseWriter, r *http.Request, cfg *Configuration) {
	rpath := cfg.RepositoryStore
	gins := &GinServer{}
	rbd := ReIndexRequest{}
	getParsedBody(r, cfg.Key, &rbd)
	log.Debugf("Got a reindex request: %+v", rbd)
	repos, err := findRepos(rpath, &rbd, gins)
	if err != nil {
		log.Debugf("Failed listing repositories: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	jobQue := make(chan IndexJob, 10)
	disp := NewDispatcher(jobQue, 3)
	disp.Run(NewWorker)
	wg := sync.WaitGroup{}

	for _, repo := range repos {
		rec := httptest.NewRecorder()
		ireq := IndexRequest{rbd.UserID, repo.FullName,
			fmt.Sprintf("%d", repo.ID)}
		data, _ := json.Marshal(ireq)
		req, _ := http.NewRequest(http.MethodPost, "/index", bytes.NewReader(data))
		wg.Add(1)
		jobQue <- IndexJob{rec, req, cfg, &wg}
	}
	wg.Wait()
	w.WriteHeader(http.StatusOK)
}

func suggest(sreq *libgin.SearchRequest, els *ESServer) ([]string, error) {
	commS, err := els.Suggest(sreq)
	defer commS.Body.Close()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(commS.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`<em>(\w+)</em>`)
	sdata := string(data)
	matches := re.FindAllStringSubmatch(string(sdata), -1)
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match[1]
	}
	result = UniqueStr(result)
	return result, nil
}

func searchCommits(sreq *libgin.SearchRequest, els *ESServer,
	result interface{}) error {
	commS, err := els.SearchCommits(sreq)
	if err != nil {
		return err
	}
	err = parseElResult(commS, &result)
	commS.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func searchBlobs(sreq *libgin.SearchRequest, els *ESServer,
	result interface{}) error {
	blobS, err := els.SearchBlobs(sreq)
	if err != nil {
		return err
	}
	err = parseElResult(blobS, &result)
	blobS.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func parseElResult(comS *http.Response, pRes interface{}) error {
	var res interface{}
	err := getParsedResponse(comS, &res)
	if err != nil {
		return err
	}
	// extract the somewhat nested search rersult
	if x, ok := res.(map[string](interface{})); ok {
		if y, ok := x["hits"].(map[string](interface{})); ok {
			err = map2struct(y["hits"], &pRes)
			return err
		}
	}
	return fmt.Errorf("could not extract elastic result:%s", res)
}
