package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	clientconfig "github.com/G-Node/gin-cli/ginclient/config"
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

	args, err := docopt.Parse(usage, nil, true, "gin-dex 0.3", false)
	if err != nil {
		log.Printf("Error while parsing command line: %v", err)
		os.Exit(-1)
	}

	if args["--debug"].(bool) {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}
	log.Debug("Starting gin-dex service")

	elURL := libgin.ReadConf("elastic_url")

	// These don't need to be configurable
	commitIndex := "commits"
	blobIndex := "blobs"

	// TODO: Remove all auth support?
	els := NewESServer(elURL, blobIndex, commitIndex, nil, nil)
	err = els.Init()
	if err != nil {
		log.Errorf("Failed to connect to elastic service: %v", err)
		os.Exit(-1)
	}
	rpath := libgin.ReadConf("repository_store")

	// TODO: Remove requirement for calling back to the GIN server
	ginURL := libgin.ReadConf("gin_url")
	gin := &GinServer{URL: ginURL}

	web, err := clientconfig.ParseWebString(ginURL)
	if err != nil {
		log.Errorf("Failed to parse GIN URL string: %v", err)
		os.Exit(-1)
	}
	srvcfg := clientconfig.ServerCfg{Web: web}
	clientconfig.AddServerConf("gin", srvcfg)

	log.Debug("Registering routes")

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

	// maxTxt: Maximum size to index for text files (in MB)
	maxTxt, err := strconv.ParseInt(libgin.ReadConfDefault("text_max", "10"), 10, 64)
	if err != nil {
		log.Printf("Error while parsing maxTxt variable: %v", err)
		maxTxt = 10
		log.Printf("Using default: %d", maxTxt)
	}
	// maxPDF: Maximum size to index for PDF files (in MB)
	maxPDF, err := strconv.ParseInt(libgin.ReadConfDefault("pdf_max", "100"), 10, 64)
	if err != nil {
		log.Printf("Error while parsing maxPDFize variable: %v", err)
		maxPDF = 100
		log.Printf("Using default: %d", maxPDF)
	}
	Setting.MaxSizeText = maxTxt
	Setting.MaxSizePDF = maxPDF

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
