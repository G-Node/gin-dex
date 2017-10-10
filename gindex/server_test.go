package gindex

import "testing"

func TestelServer(t *testing.T) {
	pw := "changeme"
	un := "elastic"
	fakeServer := ElServer{adress: "127.0.0.1:9200", uname: &un, password: &pw}
	err := IndexRepoWithPath("../tdata/repo1.git", &fakeServer, "testid")
	if err != nil {
		t.Errorf("Could  not index repo:%v", err)
	}

}
