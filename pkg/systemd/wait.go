package systemd

import (
	"fmt"
	"time"
)

// Wait waits for systemd to come up, if duration d is reached before systemd is up an error
// is returned.
func Wait(d time.Duration) error {
	start := time.Now()
	waitcmd, _ := Command(List, Options{}, "")

	for {
		if err := waitcmd.Run(); err == nil {
			// no error, it's up!
			return nil
		}
		if time.Since(start) > d {
			return fmt.Errorf("no running systemd found within %s", d)
		}
		time.Sleep(1 * time.Second)
	}
}
