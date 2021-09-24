package systemd

import (
	"fmt"
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
	Logs:    {"journalctl", usr},
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
	Logs    Operation = "logs"
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
		args = append(args, "-n")
		args = append(args, fmt.Sprintf("%d", opts.Tail))
	}
	if opts.Follow {
		args = append(args, "-f")
	}
	if !opts.Timestamps {
		args = append(args, "-o")
		args = append(args, "cat")
	} else {
		args = append(args, "-o")
		args = append(args, "short-full") // this is _not_ the default Go timestamp output
	}
	if opts.SinceSeconds > 0 {
		args = append(args, "-S")
		args = append(args, fmt.Sprintf("-%ds", opts.SinceSeconds))
	}
	if !opts.SinceTime.IsZero() {
		args = append(args, "-S")
		args = append(args, opts.SinceTime.Format(time.RFC3339))
	}

	if service != "" {
		c.Args = append(c.Args, service)
	}

	return c, nil
}

// Run runs command c and writes the response back to the user. In case of
// logging this will be a streaming response.
func Run(c *exec.Cmd, stream bool, w http.ResponseWriter) error {
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %s", c, err)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return nil
}
