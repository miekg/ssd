package systemd

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
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

func ParseOptions(opts url.Values) (Options, error) {
	o := Options{}
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
