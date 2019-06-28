package main

import (
	"github.com/G-Node/gig"
	log "github.com/sirupsen/logrus"
)

// IndexRepoWithPath walks a repository at a given path and submits it to the index
func IndexRepoWithPath(cfg *Configuration, path, ref string, repoid string, reponame string) error {
	log.Infof("Start indexing repository with path: %s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		log.Errorf("Could not open repository: %v", err)
		return err
	}
	log.Debugf("Opened repository")
	commits, err := rep.WalkRef(ref,
		func(comitID gig.SHA1) bool {
			res, ierr := cfg.Elasticsearch.HasCommit("commits", GetIndexCommitId(comitID.String(), repoid))
			if ierr != nil {
				log.Errorf("Could not query commit index: %v", err)
				return false
			}
			return !res
		})
	if err != nil {
		log.Errorf("Refwalk for repository %s failed: %v", path, err)
		return err
	}
	log.Infof("Found %d commits", len(commits))

	// TODO: Fix error handling in loop
	for commitid, commit := range commits {
		err = indexCommit(cfg, commit, repoid, commitid, rep, path, reponame, cfg.Elasticsearch.HasBlob)
	}
	return err
}

// reIndexRepoWithPath walks a repository at a given path and resubmits it to the index
func reIndexRepoWithPath(cfg *Configuration, path, ref string, repoid string, reponame string) error {
	log.Infof("Start indexing repository with path: %s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		log.Errorf("Could not open repository: %v", err)
		return err
	}
	log.Debugf("Opened repository")
	commits, err := rep.WalkRef(ref,
		func(comitID gig.SHA1) bool {
			return true
		})
	if err != nil {
		log.Errorf("Refwalk for repository %s failed: %v", path, err)
		return err
	}
	log.Infof("Found %d commits", len(commits))

	blobs := make(map[gig.SHA1]bool)
	// TODO: Fix error handling in loop
	for commitid, commit := range commits {
		err = indexCommit(cfg, commit, repoid, commitid, rep, path, reponame,
			func(indexName string, id gig.SHA1) (bool, error) {
				if !blobs[id] {
					blobs[id] = true
					return false, nil
				}
				return true, nil
			})
		if err != nil {
			log.Errorf("Indexing for repository %s failed: %v", path, err)
		}
	}
	return nil
}

func indexCommit(cfg *Configuration, commit *gig.Commit, repoid string, commitid gig.SHA1, rep *gig.Repository, path string, reponame string, indexBlob func(string, gig.SHA1) (bool, error)) error {
	err := NewCommitFromGig(commit, repoid, reponame, commitid).AddToIndex(cfg.Elasticsearch, "commits", commitid)
	if err != nil {
		log.Printf("Indexing commit failed: %v", err)
	}
	blobs := make(map[gig.SHA1]*gig.Blob)
	rep.GetBlobsForCommit(commit, blobs)
	// TODO: Fix error handling in loop
	for blid, blob := range blobs {
		log.Debugf("Blob %s has Size: %d", blid, blob.Size())
		hasBlob, err := indexBlob("blobs", GetIndexBlobId(blid.String(), repoid))
		if err != nil {
			log.Errorf("Could not query for blob: %v", err)
			return err
		}
		if !hasBlob {
			bpath, _ := GetBlobPath(blid.String(), commitid.String(), path)
			err = NewBlobFromGig(blob, repoid, blid, commitid.String(), bpath, reponame).AddToIndexTimeout(cfg, blid)
			if err != nil {
				log.Debugf("Indexing blob failed: %v", err)
			}
		} else {
			log.Debugf("Blob there :%s", blid)
		}
	}
	return nil
}
