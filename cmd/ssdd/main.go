package main

import (
	"flag"
	"fmt"
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

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			// In case of error, nothing has been written yet.
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err.Error())
		}
	})
	// TODO: metrics

	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting service on port %s", *flgPort)
	err := http.ListenAndServe(":"+*flgPort, nil)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
