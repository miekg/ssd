package main

import (
	"fmt"
	"log"
	"miekg/ssd/pkg/systemd"
	"net/http"
	"path"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	p := path.Clean(r.URL.Path)
	pcs := strings.Split(p, "/")
	if len(pcs) == 0 {
		return fmt.Errorf("URL is too short: %s", p)
	}
	// url start with a / so first element should be empty
	pcs = pcs[1:]

	if len(pcs) < 2 {
		return fmt.Errorf("URL is too short: %s", p)
	}
	if len(pcs) > 3 {
		return fmt.Errorf("URL has crap at the end: %s", p)
	}
	if pcs[0] != "s" {
		return fmt.Errorf("URL needs to start with /s: %s", p)
	}
	operation := pcs[1]
	service := ""
	if len(pcs) == 3 {
		service = pcs[2]
	}

	opts, err := systemd.ParseOptions(r.Form)
	if err != nil {
		return err
	}

	cmd, err := systemd.Command(systemd.Operation(operation), opts, service)
	if err != nil {
		return err
	}
	stream := false
	switch systemd.Operation(operation) {
	case systemd.Logs:
		stream = true
	case systemd.List:
		// nothing
	default:
		if service == "" {
			return fmt.Errorf("this operation %s requires a service", operation)
		}
	}
	log.Printf("Running command %q initiated from %q", cmd, r.RemoteAddr)
	return systemd.Run(cmd, stream, w)
}
