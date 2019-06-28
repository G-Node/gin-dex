package main

import (
	"fmt"
	"net/http"
	"os"

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

	cfg := loadconfig()

	log.Debug("Registering routes")
	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, cfg)
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler(w, r, cfg)
	})
	http.HandleFunc("/suggest", func(w http.ResponseWriter, r *http.Request) {
		suggestHandler(w, r, cfg.Elasticsearch)
	})
	http.HandleFunc("/reindex", func(w http.ResponseWriter, r *http.Request) {
		reIndexHandler(w, r, cfg)
	})

	fmt.Printf("Listening for connections on port %d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
