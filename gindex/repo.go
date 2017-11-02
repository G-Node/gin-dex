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
		res, err := serv.HasCommit("commits", GetIndexCommitId(comitID.String(), repoid))
		if err != nil {
			log.Errorf("Could not querry commit index: %v", err)
			return false
		}
		return !res
	})
	log.Infof("Found %d commits", len(commits))

	for commitid, commit := range commits {
		err := NewCommitFromGig(commit, repoid, commitid).AddToIndex(serv, "commits", commitid)
		if err != nil {
			log.Printf("Indexing commit failed:%+v", err)
		}
		blobs := make(map[gig.SHA1]*gig.Blob)
		rep.GetBlobsForCommit(commit, blobs)
		for blid, blob := range blobs {
			log.Debugf("Blob has Size:%d", blob.Size())
			hasBlob, err := serv.HasBlob("blobs", GetIndexBlobId(blid.String(), repoid))
			if err != nil {
				log.Errorf("Could not querry for blob: %+v", err)
				return err
			}
			if !hasBlob {
				err = NewBlobFromGig(blob, repoid, blid, commitid.String()).AddToIndex(serv, "blobs", path, blid)
				if err != nil {
					log.Debugf("Indexing blob failed: %+v", err)
				}
			} else {
				log.Debugf("Blob there :%s", blid)
			}
		}
	}
	return nil
}
