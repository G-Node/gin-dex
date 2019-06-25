package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/G-Node/gig"
	log "github.com/sirupsen/logrus"
)

// ESServer defines an ElasticSearch server to be used for indexing and search.
type ESServer struct {
	address  string
	uname    *string
	password *string
	blindex  string
	coindex  string
}

// NewESServer initialises and returns a new ESServer configuration.
func NewESServer(address, blindex, coindex string, uname, password *string) *ESServer {
	return &ESServer{address: address, uname: uname, password: password, blindex: blindex, coindex: coindex}
}

// Init initialises indexes and mappings in the ElasticSearch server.
func (el *ESServer) Init() error {
	// TODO: Check if mappings already exist and skip
	// create Blob mapping
	log.Debugf("Connecting to %s", el.address)
	addr := fmt.Sprintf("%s/%s/", el.address, el.blindex)
	log.Debugf("Adding blob mapping to %s", addr)
	req, err := http.NewRequest("PUT", addr, bytes.NewReader([]byte(BLOB_MAPPING)))
	if err != nil {
		return err
	}
	resp, err := el.elasticRequest(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		log.Infof("Blob Mapping not created: %d, %s", resp.StatusCode, data)
	}

	// create Commit mapping
	addr = fmt.Sprintf("%s/%s/", el.address, el.coindex)
	log.Debugf("Adding commit mapping to %s", addr)
	req, err = http.NewRequest("PUT", addr, bytes.NewReader([]byte(COMMIT_MAPPING)))
	if err != nil {
		return err
	}
	resp, err = el.elasticRequest(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		log.Infof("Commit Mapping not created: %d, %s", resp.StatusCode, data)
	}
	return nil
}

func (el *ESServer) Index(index, doctype string, data []byte, id gig.SHA1) (*http.Response, error) {
	addr := fmt.Sprintf("%s/%s/%s/%s", el.address, index, doctype, id.String())
	req, err := http.NewRequest("POST", addr, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ESServer) elasticRequest(req *http.Request) (*http.Response, error) {
	if el.uname != nil {
		req.SetBasicAuth(*el.uname, *el.password)
	}
	req.Header.Set("Content-Type", "application/json")
	cl := http.Client{}
	return cl.Do(req)
}

func (el *ESServer) HasCommit(index string, commitId gig.SHA1) (bool, error) {
	addr := fmt.Sprintf("%s/%s/commit/%s", el.address, el.coindex, commitId)
	return el.Has(addr)
}

func (el *ESServer) HasBlob(index string, blobId gig.SHA1) (bool, error) {
	addr := fmt.Sprintf("%s/%s/blob/%s", el.address, el.blindex, blobId)
	return el.Has(addr)
}

func (el *ESServer) Has(addr string) (bool, error) {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return false, err
	}
	resp, err := el.elasticRequest(req)
	if err != nil {
		return false, err
	}
	bdy, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var res struct{ Found bool }
	err = json.Unmarshal(bdy, &res)
	if err != nil {
		log.WithError(err)
		return false, err
	}
	return res.Found, nil
}

func (el *ESServer) search(query, addr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", addr, bytes.NewReader([]byte(query)))
	if err != nil {
		log.Errorf("Could not form search query: %v", err)
		log.Errorf("Formatted query was: %s", query)
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ESServer) SearchBlobs(query string, okRepos []string, searchType int64) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %v", err)
		return nil, err
	}
	var formatted_query string
	switch searchType {
	case SEARCH_FUZZY:
		formatted_query = fmt.Sprintf(BLOB_FUZ_QUERY, query, string(repos))
	case SEARCH_WILDCARD:
		formatted_query = fmt.Sprintf(BLOB_WC_QUERY, strings.ToLower(query), string(repos))
	case SEARCH_QUERY:
		formatted_query = fmt.Sprintf(BLOB_QString_QUERY, query, string(repos))
	default:
		formatted_query = fmt.Sprintf(BLOB_QUERY, query, string(repos))
	}

	addr := fmt.Sprintf("%s/%s/_search", el.address, el.blindex)
	return el.search(formatted_query, addr)
}

func (el *ESServer) SearchCommits(query string, okRepos []string) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %v", err)
		return nil, err
	}
	formattedQuery := fmt.Sprintf(COMMIT_QUERY, query, string(repos))
	addr := fmt.Sprintf("%s/%s/_search", el.address, el.coindex)
	return el.search(formattedQuery, addr)
}

func (el *ESServer) Suggest(query string, okRepos []string) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %v", err)
		return nil, err
	}
	formatted_query := fmt.Sprintf(SUGGEST_QUERY, query, string(repos))
	addr := fmt.Sprintf("%s/%s/_search", el.address, el.blindex)
	return el.search(formatted_query, addr)
}

var BLOB_QUERY = `{
"from" : 0, "size" : 20,
	"_source": ["Oid","GinRepoName","FirstCommit","Path"],
	"query": {
		"bool": {
			"must": {
				"multi_match": {
					"query": "%s"
				}
			},
			"filter": {
				"terms": {
					"GinRepoId" : %s
				}
			}
		}
	},
	"highlight" : {
		"fields" : [
			{"Content" : {
				"fragment_size" : 100,
				"number_of_fragments" : 10,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
			}
			}
		]
	}
}`

var BLOB_FUZ_QUERY = `{
"from" : 0, "size" : 20,
		"_source": ["Oid","GinRepoName","FirstCommit","Path"],
		"query": {
		"bool": {
			"must": {
			"fuzzy": {
				"_all":"%s"
			}
			},
			"filter": {
			"terms": {
				"GinRepoId" : %s
			}
		}
		}
	},
	"highlight" : {
		"fields" : [
			{"Content" : {
				"fragment_size" : 100,
				"number_of_fragments" : 10,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`

var BLOB_WC_QUERY = `{
"from" : 0, "size" : 20,
		"_source": ["Oid","GinRepoName","FirstCommit","Path"],
		"query": {
		"bool": {
			"must": {
			"wildcard": {
				"_all":"%s"
			}
			},
			"filter": {
			"terms": {
				"GinRepoId" : %s
			}
		}
		}
	},
	"highlight" : {
		"fields" : [
			{"Content" : {
				"fragment_size" : 100,
				"number_of_fragments" : 10,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`

var BLOB_QString_QUERY = `{
"from" : 0, "size" : 20,
		"_source": ["Oid","GinRepoName","FirstCommit","Path"],
		"query": {
		"bool": {
			"must": {
			"query_string": {
				"default_field" : "Content",
				"query":"%s"
			}
			},
			"filter": {
			"terms": {
				"GinRepoId" : %s
			}
		}
		}
	},
	"highlight" : {
		"fields" : [
			{"Content" : {
				"fragment_size" : 100,
				"number_of_fragments" : 10,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`

var COMMIT_QUERY = `{
"from" : 0, "size" : 20,
		"_source": ["Oid","GinRepoName","FirstCommit","Path"],
		"query": {
		"bool": {
			"must": {
			"match": {
				"_all": "%s"
			}
			},
			"filter": {
			"terms": {
				"GinRepoId" : %s
			}
			}
		}
	},
	"highlight" : {
		"fields" : [
			{"Message" : {
				"fragment_size" : 50,
				"number_of_fragments" : 3,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			},
			{"GinRepoName" : {
				"fragment_size" : 50,
				"number_of_fragments" : 3,
				"fragmenter": "span",
				"require_field_match":false,
				"pre_tags" : ["<b>"],
				"post_tags" : ["</b>"]
				}
			}
		]
	}
}`

var SUGGEST_QUERY = `{
"from": 0,
	"size": 20,
	"_source": [
	""
	],
	"query": {
	"bool": {
		"must": {
		"match_phrase_prefix": {
			"Content": {
			"query": "%s",
			"max_expansions": 10
			}
		}
		},
		"filter": {
		"terms": {
			"GinRepoId": %s
		}
		}
	}
	},
	"highlight": {
	"fields": {
		"Content": {}
	},
	"boundary_scanner": "word"
	}
}`

var BLOB_MAPPING = `{
"mappings": {
	"blob": {
		"properties": {
		"Content": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"FirstCommit": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"GinRepoId": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"GinRepoName": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"Id": {
			"type": "long"
		},
		"IndexingTime": {
			"type": "date"
		},
		"Oid": {
			"type": "long"
		},
		"Path": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		}
		}
	}
	}
}`

var COMMIT_MAPPING = `{
"mappings": {
	"commit": {
		"properties": {
		"Author": {
			"properties": {
			"Date": {
				"type": "date"
			},
			"Email": {
				"type": "text",
				"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256
				}
				}
			},
			"Name": {
				"type": "text",
				"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256
				}
				}
			},
			"Offset": {
				"type": "object"
			}
			}
		},
		"Committer": {
			"properties": {
			"Date": {
				"type": "date"
			},
			"Email": {
				"type": "text",
				"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256
				}
				}
			},
			"Name": {
				"type": "text",
				"fields": {
				"keyword": {
					"type": "keyword",
					"ignore_above": 256
				}
				}
			},
			"Offset": {
				"type": "object"
			}
			}
		},
		"GPGSig": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"GinRepoId": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"GinRepoName": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"IndexingTime": {
			"type": "date"
		},
		"Message": {
			"type": "text",
			"fields": {
			"keyword": {
				"type": "keyword",
				"ignore_above": 256
			}
			}
		},
		"Oid": {
			"type": "long"
		},
		"Parent": {
			"type": "long"
		},
		"Tree": {
			"type": "long"
		}
		}
	}
	}
}`
