package gindex

import (
	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

func IndexRepoWithPath(path string, serv *ElServer, repoid string) error {
	log.Info("Start indexing a repository with path:%s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		return err
	}
	log.Info("Did open repo")
	commits, err := rep.WalkRef("tag1", func(comitID gig.SHA1) bool {
		res, err := serv.HasCommit("commits", comitID)
		if err != nil {
			log.Printf("Could not querry commit index: %v", err)
			return false
		}
		return !res
	})
	log.Infof("Found n commits: %d", len(commits))
	if err != nil {
		return err
	}
	for _, commit := range commits {
		err := NewCommitFromGig(commit, repoid).AddToIndex(serv, "commits")
		blobs := make(map[gig.SHA1]*gig.Blob)
		rep.GetBlobsForCommit(commit, blobs)
		for blid, blob := range blobs {
			hasBlob, err := serv.HasBlob("blobs", blid)
			if err != nil {
				return err
			}
			if !hasBlob {
				NewBlobFromGig(blob, repoid).AddToIndex(serv, "blobs")
			}
		}
		if err != nil {
			log.Printf("Big problem2")
		}
	}
	return nil
}
