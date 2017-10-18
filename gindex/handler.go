package gindex

import "net/http"

// Handler for Index requests
func IndexH(w http.ResponseWriter, r *http.Request, els *ElServer) {
	w.WriteHeader(http.StatusOK)
}

// Handler for Search requests
func SearchH(w http.ResponseWriter, r *http.Request, els *ElServer) {
	w.WriteHeader(http.StatusOK)
}
