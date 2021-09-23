package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"miekg/ssd/pkg/authn"
	"miekg/ssd/pkg/authz"
	"miekg/ssd/pkg/systemd"
)

var (
	flgPort  = flag.String("p", "9000", "port to use")
	flgAuth  = flag.Bool("auth", true, "authenticate the request")
	flgUsers = flag.String("users", "", "file to load with users allowed to perform command")
)

func main() {
	flag.Parse()
	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		user := ""
		if *flgAuth {
			var err error
			user, err = authn.Identify(r)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				return
			}
			if user == "" {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "no user found")
				return
			}
			ok, err := authz.IsAllowed(user, *flgUsers)
			if err != nil {
				// In case of error, nothing has been written yet.
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err.Error())
				return
			}
			if !ok {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				return
			}
		}

		err := handler(w, r, user)
		if err != nil {
			// In case of error, nothing has been written yet.
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err.Error())
			return
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
