package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"miekg/ssd/pkg/systemd"
)

var (
	flgPort = flag.String("p", "9000", "port to use")
)

func main() {
	flag.Parse()
	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	// start webserver, add routers
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "Hello world\n")
	})
	// metrics

	err := http.ListenAndServe(":"+"*flgPort", nil)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
