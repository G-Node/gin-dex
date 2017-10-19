package main

import (
	"github.com/docopt/docopt-go"
	"os"
	"github.com/G-Node/gin-dex/gindex"
	"net/http"
	log  "github.com/Sirupsen/logrus"
)

func main() {
	usage := `gin-dex.
Usage:
  gin-dex [--eladress=<eladress> --eluser=<eluser> --elpw=<elpw> --rpath=<rpath> --port=<port> --debug]

Options:
  --eladress=<eladress>           Adress of the elastic server [default: http://localhost:9200]
  --eluser=<eluser>               Elastic user [default: elastic]
  --elpw=<elpw>                   Elastic password [default: changeme]
  --port=<port>                   Server port [default: 8099]
  --debug                         Whether debug messages shall be printed
 `

	args, err := docopt.Parse(usage, nil, true, "gin-dex0.1a", false)
	if err != nil {
		log.Printf("Error while parsing command line: %+v", err)
		os.Exit(-1)
	}

	els := &gindex.ElServer{args["--eladress"].(string), &args["--eladress"].(string),
		&args["--eladress"].(string)}

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		gindex.IndexH(w, r, els, &args["--bpath"].(string))
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		gindex.SearchH(w, r, els)
	})

	if args["--debug"].(bool) {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}
	log.Fatal(http.ListenAndServe(":"+args["--port"].(string), nil))
}