package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/G-Node/libgin/libgin"
	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
)

func main() {
	usage := `gin-dex.
Usage:
  gin-dex [--debug]

Options:
  --debug                         Print debug messages
`

	args, err := docopt.Parse(usage, nil, true, "gin-dex0.2", false)
	if err != nil {
		log.Printf("Error while parsing command line: %v", err)
		os.Exit(-1)
	}

	if args["--debug"].(bool) {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}
	log.Debug("Starting gin-dex service")

	elURL := libgin.ReadConf("elurl")

	// These don't need to be configurable
	commitIndex := "commits"
	blobIndex := "blobs"

	// TODO: Remove all auth support?
	els := NewElServer(elURL, blobIndex, commitIndex, nil, nil)
	err = els.Init()
	if err != nil {
		log.Errorf("Failed to connect to elastic service: %v", err)
		os.Exit(-1)
	}
	rpath := libgin.ReadConf("rpath")

	// TODO: Remove requirement for calling back to the GIN server
	gin := &GinServer{URL: "https://gin.g-node.org"}

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		IndexH(w, r, els, &rpath)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		SearchH(w, r, els, gin)
	})

	http.HandleFunc("/suggest", func(w http.ResponseWriter, r *http.Request) {
		SuggestH(w, r, els, gin)
	})

	http.HandleFunc("/reindex", func(w http.ResponseWriter, r *http.Request) {
		ReindexH(w, r, els, gin, &rpath)
	})

	// txtMs: Maximum size to index for text files (in MB)
	txtMs, err := strconv.ParseInt(libgin.ReadConfDefault("txtMSize", "10"), 10, 64)
	if err != nil {
		log.Printf("Error while parsing txtMs variable: %v", err)
		txtMs = 10
		log.Printf("Using default: %d", txtMs)
	}
	// txtMs: Maximum size to index for PDF files (in MB)
	pdfMs, err := strconv.ParseInt(libgin.ReadConfDefault("pdfMSize", "100"), 10, 64)
	if err != nil {
		log.Printf("Error while parsing pdfMsize variable: %v", err)
		pdfMs = 100
		log.Printf("Using default: %d", pdfMs)
	}
	Setting.TxtMSize = txtMs
	Setting.PdfMSize = pdfMs

	// timeout for adding contents to index (in seconds)
	timeout, err := strconv.ParseInt(libgin.ReadConfDefault("timeout", "60"), 10, 64)
	if err != nil {
		log.Printf("Error while parsing timeout variable: %v", err)
		timeout = 60
		log.Printf("Using default: %d", timeout)
	}
	Setting.Timeout = timeout

	port := libgin.ReadConf("port")
	fmt.Printf("Listening for connections on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
