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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flgPort  = flag.String("p", "9999", "port to use")
	flgAuth  = flag.Bool("auth", true, "authenticate the request")
	flgUsers = flag.String("users", "", "file to load with users allowed to perform command")
)

func main() {
	flag.Parse()
	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	var (
		requests = promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ssdd",
			Name:      "requests_total",
			Help:      "Counter of incoming requests.",
		})
		errors = promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ssdd",
			Name:      "errors_total",
			Help:      "Counter of requests ending in failure.",
		})
	)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		requests.Inc()
		user := ""
		if *flgAuth {
			var err error
			user, err = authn.Identify(r)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				errors.Inc()
				return
			}
			if user == "" {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "no user found")
				errors.Inc()
				return
			}
			ok, err := authz.IsAllowed(user, *flgUsers)
			if err != nil {
				// In case of error, nothing has been written yet.
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err.Error())
				errors.Inc()
				return
			}
			if !ok {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				errors.Inc()
				return
			}
		}
		if user == "" {
			log.Printf("Running command on behalve of an anonymous user, from %q", r.RemoteAddr)
		} else {
			log.Printf("Running command on behalve of user %q, from %q", user, r.RemoteAddr)
		}

		err := handler(w, r)
		if err != nil {
			log.Printf("Command failed to run: %s", err)
			// In case of error, nothing has been written yet.
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err.Error())
			errors.Inc()
			return
		}
	})

	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting service on port %s", *flgPort)
	err := http.ListenAndServe(":"+*flgPort, nil)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
