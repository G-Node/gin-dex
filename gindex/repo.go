package gindex

import (
	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

func IndexRepoWithPath(path, ref string, serv *ElServer, repoid string) error {
	log.Info("Start indexing repository with path: %s", path)
	rep, err := gig.OpenRepository(path)
	if err != nil {
		log.Errorf("Could not open repository: %+v", err)
		return err
	}
	log.Debugf("Opened repository")
	commits, err := rep.WalkRef(ref, func(comitID gig.SHA1) bool {
		res, err := serv.HasCommit("commits", comitID)
		if err != nil {
			log.Errorf("Could not querry commit index: %v", err)
			return false
		}
		return !res
	})
	log.Infof("Found %d commits", len(commits))

	for id, commit := range commits {
		err := NewCommitFromGig(commit, repoid).AddToIndex(serv, "commits", id)
		if err != nil {
			log.Printf("Indexing commit failed:%+v", err)
		}
		blobs := make(map[gig.SHA1]*gig.Blob)
		rep.GetBlobsForCommit(commit, blobs)
		for blid, blob := range blobs {
			log.Debugf("Blob has Size:%d", blob.Size())
			hasBlob, err := serv.HasBlob("blobs", blid)
			if err != nil {
				log.Errorf("Could not querry for blob: %+v", err)
				return err
			}
			if !hasBlob {
				err = NewBlobFromGig(blob, repoid).AddToIndex(serv, "blobs", path, blid)
				if err != nil {
					log.Debugf("Indexing blob failed: %+v", err)
				}
			}
		}
	}
	return nil
}
