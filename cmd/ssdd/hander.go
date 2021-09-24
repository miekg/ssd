package main

import (
	"fmt"
	"io"
	"log"
	"miekg/ssd/pkg/systemd"
	"net/http"
	"path"
	"strings"
	"time"
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

	opts, err := systemd.ParseOptions(r.URL.Query())
	if err != nil {
		return err
	}

	cmd, err := systemd.Command(systemd.Operation(operation), opts, service)
	if err != nil {
		return err
	}
	switch systemd.Operation(operation) {
	case systemd.Log:
		// nothing
	case systemd.List:
		// nothing
	default:
		if service == "" {
			return fmt.Errorf("this operation %s requires a service", operation)
		}
	}
	log.Printf("Running command %q initiated from %q", cmd, r.RemoteAddr)

	if !opts.Follow {
		return systemd.Run(cmd, w)

	}

	// Should only be the case for journalctl, but we don't really care.
	rc, cancel, err := systemd.RunPipe(cmd)
	if err != nil {
		return err
	}
	defer rc.Close()
	defer cancel()

	// If in follow mode, follow until interrupted.
	untilTime := make(chan time.Time, 1)
	errChan := make(chan error, 1)

	go func(w io.Writer, errChan chan error) {
		err := systemd.Follow(untilTime, rc, w)
		errChan <- err
	}(systemd.FlushWriter(w), errChan)

	// Stop following logs if request context is completed.
	select {
	case err := <-errChan:
		return err
	case <-r.Context().Done():
		close(untilTime)
	}
	return nil
}
