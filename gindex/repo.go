package gindex

import (
	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

func IndexRepoWithPath(path string, serv *ElServer) error {
	rep, err := gig.OpenRepository(path)
	if err != nil {
		return err
	}
	commits, err := rep.WalkRef("master", func(comitID gig.SHA1) bool {
		res, err := serv.HasCommit("commits", comitID)
		if err != nil {
			log.Printf("Could not querry commit index: %v", err)
			return false
		}
		return res
	})
	if err != nil {
		return err
	}
	for _, commit := range commits{
		NewCommitFromGig(commit).AddToIndex(serv, "commits")
	}
	return nil
}
