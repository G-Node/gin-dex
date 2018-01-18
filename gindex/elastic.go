package gindex

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"

	"io/ioutil"

	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type ElServer struct {
	adress   string
	uname    *string
	password *string
	blindex  string
	coindex  string
}

func NewElServer(adress, blindex, coindex string, uname, password *string) *ElServer {
	return &ElServer{adress: adress, uname: uname, password: password, blindex: blindex, coindex: coindex}
}

func (el *ElServer) Init() error {
	// create Blob mapping
	adrr := fmt.Sprintf("%s/%s/", el.adress, el.blindex)
	req, err := http.NewRequest("PUT", adrr, bytes.NewReader([]byte(BLOB_MAPPING)))
	if err != nil {
		return err
	}
	resp, err := el.elasticRequest(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		log.Infof("Blob Mapping not created:%d, %s", resp.StatusCode, data)
	}

	// create Commit mapping
	adrr = fmt.Sprintf("%s/%s/", el.adress, el.coindex)
	req, err = http.NewRequest("PUT", adrr, bytes.NewReader([]byte(COMMIT_MAPPING)))
	if err != nil {
		return err
	}
	resp, err = el.elasticRequest(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		log.Infof("Commit Mapping not created:%d, %s", resp.StatusCode, data)
	}
	return nil
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
	return cl.Do(req)
}

func (el *ElServer) HasCommit(index string, commitId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/%s/commit/%s", el.adress, el.coindex, commitId)
	return el.Has(adrr)
}

func (el *ElServer) HasBlob(index string, blobId gig.SHA1) (bool, error) {
	adrr := fmt.Sprintf("%s/%s/blob/%s", el.adress, el.blindex, blobId)
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
	resp.Body.Close()
	var res struct{ Found bool }
	err = json.Unmarshal(bdy, &res)
	if err != nil {
		log.WithError(err)
		return false, err
	}
	return res.Found, nil
}

func (el *ElServer) search(querry, adrr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", adrr, bytes.NewReader([]byte(querry)))
	if err != nil {
		log.Errorf("Could not form search query:%+v", err)
		log.Errorf("Formatted query was:%s", querry)
		return nil, err
	}
	return el.elasticRequest(req)
}

func (el *ElServer) SearchBlobs(querry string, okRepos []string, searchType int64) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %+v", err)
		return nil, err
	}
	var formatted_querry string
	switch searchType {
	case SEARCH_FUZZY:
		formatted_querry = fmt.Sprintf(BLOB_FUZ_QUERRY, querry, string(repos))
	case SEARCH_WILDCARD:
		formatted_querry = fmt.Sprintf(BLOB_WC_QUERRY, strings.ToLower(querry), string(repos))
	case SEARCH_QUERRY:
		formatted_querry = fmt.Sprintf(BLOB_QString_QUERRY, querry, string(repos))
	default:
		formatted_querry = fmt.Sprintf(BLOB_QUERRY, querry, string(repos))
	}

	adrr := fmt.Sprintf("%s/%s/_search", el.adress, el.blindex)
	return el.search(formatted_querry, adrr)
}

func (el *ElServer) SearchCommits(querry string, okRepos []string) (*http.Response, error) {
	//implement the passing of the repo ids
	repos, err := json.Marshal(okRepos)
	if err != nil {
		log.Errorf("Could not marshal okRepos: %+v", err)
		return nil, err
	}
	formatted_querry := fmt.Sprintf(COMMIT_QUERRY, querry, string(repos))
	adrr := fmt.Sprintf("%s/%s/_search", el.adress, el.coindex)
	return el.search(formatted_querry, adrr)
}

var BLOB_QUERRY = `{
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

var BLOB_FUZ_QUERRY = `{
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

var BLOB_WC_QUERRY = `{
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

var BLOB_QString_QUERRY = `{
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

var COMMIT_QUERRY = `{
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

var BLOB_MAPPING = `{
  "mappings": {
    "blob": {
      "_all": {
        "enabled": true
      },
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