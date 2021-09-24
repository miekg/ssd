package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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
	flgHelp  = flag.Bool("h", false, "show help")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println(helpMsg)
	}

	flag.Parse()

	if *flgHelp {
		flag.Usage()
		return
	}

	if err := systemd.Wait(5 * time.Minute); err != nil {
		log.Fatal(err)
	}

	var (
		promRequests = promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ssdd",
			Name:      "requests_total",
			Help:      "Counter of incoming requests.",
		})
		promErrors = promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ssdd",
			Name:      "promErrors_total",
			Help:      "Counter of requests ending in failure.",
		})
	)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/h", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(helpMsg))
	})

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		promRequests.Inc()
		user := ""
		if *flgAuth {
			var err error
			user, err = authn.Identify(r)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				promErrors.Inc()
				return
			}
			if user == "" {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "no user found")
				promErrors.Inc()
				return
			}
			ok, err := authz.IsAllowed(user, *flgUsers)
			if err != nil {
				// In case of error, nothing has been written yet.
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err.Error())
				promErrors.Inc()
				return
			}
			if !ok {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, err.Error())
				promErrors.Inc()
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
			promErrors.Inc()
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

const helpMsg = `The following endpoints are exposed:

* /h - you are reading this now
* /s - execute systemd commands

The "REST" API for /s is a follows: /s/OPERATION[/SERVICE]?OPTIONS

Where OPERATION can be:



And SERVICE is the unit file you want to act up on - usually a service.

The OPTIONS are:

`
