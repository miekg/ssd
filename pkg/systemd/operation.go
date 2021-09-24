package systemd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"time"
)

const (
	sysctl = "systemctl"
	usr    = "--user"
)

// OperationToCommand contains a mapping from the operation to the actual
// command being run.
var OperationToCommand = map[Operation][]string{
	List:    {sysctl, usr, "list-units"},
	Cat:     {sysctl, usr, "cat"},
	Start:   {sysctl, usr, "start"},
	Status:  {sysctl, usr, "status"},
	Stop:    {sysctl, usr, "stop"},
	Reload:  {sysctl, usr, "reload"},
	Restart: {sysctl, usr, "restart"},
	Log:     {"journalctl", usr},
}

// Operations specifies a systemd/journald operation.
type Operation string

const (
	List    Operation = "list"
	Cat     Operation = "cat"
	Start   Operation = "start"
	Status  Operation = "status"
	Stop    Operation = "stop"
	Reload  Operation = "reload"
	Restart Operation = "restart"
	Log     Operation = "log"
)

// Options are extra options that can be given to operations.
type Options struct {
	Tail         int
	LimitBytes   int
	Timestamps   bool
	Follow       bool
	Previous     bool
	SinceSeconds int
	SinceTime    time.Time
}

func ParseOptions(q url.Values) (Options, error) {
	fmt.Printf("%v\n", q)
	o := Options{}
	var err error

	if tailLines := q.Get("tailLines"); tailLines != "" {
		o.Tail, err = strconv.Atoi(tailLines)
		if err != nil {
			return o, err
		}
		if o.Tail < 0 {
			return o, fmt.Errorf("tailLines can't be < 0 ")
		}
	}
	if follow := q.Get("follow"); follow != "" {
		o.Follow, err = strconv.ParseBool(follow)
		if err != nil {
			return o, err
		}
	}
	if limitBytes := q.Get("limitBytes"); limitBytes != "" {
		o.LimitBytes, err = strconv.Atoi(limitBytes)
		if err != nil {
			return o, err
		}
	}
	if previous := q.Get("previous"); previous != "" {
		o.Previous, err = strconv.ParseBool(previous)
		if err != nil {
			return o, err
		}
	}
	if sinceSeconds := q.Get("sinceSeconds"); sinceSeconds != "" {
		o.SinceSeconds, err = strconv.Atoi(sinceSeconds)
		if err != nil {
			return o, err
		}
	}
	if sinceTime := q.Get("sinceTime"); sinceTime != "" {
		o.SinceTime, err = time.Parse(time.RFC3339, sinceTime)
		if err != nil {
			return o, err
		}
	}
	if timestamps := q.Get("timestamps"); timestamps != "" {
		o.Timestamps, err = strconv.ParseBool(timestamps)
		if err != nil {
			return o, err
		}
	}

	return o, nil
}

func Command(op Operation, opts Options, service string) (*exec.Cmd, error) {
	fmt.Printf("%v\n", opts)
	args, ok := OperationToCommand[op]
	if !ok {
		return nil, fmt.Errorf("no command found for %s", op)
	}
	// as OperationToCommand returns a slice which we want/may append to, we need to copy
	// it, so we own the slice and the underlaying array
	cmdline := make([]string, len(args))
	copy(cmdline, args)

	c := exec.Command(cmdline[0], cmdline[1:]...)

	if opts.Tail > 0 {
		c.Args = append(c.Args, "-n")
		c.Args = append(c.Args, fmt.Sprintf("%d", opts.Tail))
	}
	if opts.Follow {
		c.Args = append(c.Args, "-f")
	}
	if !opts.Timestamps {
		c.Args = append(c.Args, "-o")
		c.Args = append(c.Args, "cat")
	} else {
		c.Args = append(c.Args, "-o")
		c.Args = append(c.Args, "short-full") // this is _not_ the default Go timestamp output
	}
	if opts.SinceSeconds > 0 {
		c.Args = append(c.Args, "-S")
		c.Args = append(c.Args, fmt.Sprintf("-%ds", opts.SinceSeconds))
	}
	if !opts.SinceTime.IsZero() {
		c.Args = append(c.Args, "-S")
		c.Args = append(c.Args, opts.SinceTime.Format(time.RFC3339))
	}

	if service != "" {
		if op == Log {
			c.Args = append(c.Args, "-u")
		}
		c.Args = append(c.Args, service)
	}

	return c, nil
}

// Run runs command c and writes the response back to the user. In case of
// logging this will be a streaming response.
func Run(c *exec.Cmd, w http.ResponseWriter) error {
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %s", c, err)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return nil
}

// RunPipe runs command c and attaches a pipe and returns that as an io.ReadCloser.
func RunPipe(c *exec.Cmd) (io.ReadCloser, func() error, error) {
	cancel := func() error { return nil }

	p, err := c.StdoutPipe()
	if err != nil {
		return nil, cancel, err
	}

	if err := c.Start(); err != nil {
		return nil, cancel, err
	}

	cancel = func() error {
		go func() {
			if err := c.Wait(); err != nil {
				log.Printf("wait for %q failed: %s", c, err)
			}
		}()
		return c.Process.Kill()
	}

	return p, cancel, nil
}
