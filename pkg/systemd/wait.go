package systemd

import (
	"fmt"
	"log"
	"time"
)

// Wait waits for systemd to come up, if duration d is reached before systemd is up an error
// is returned.
func Wait(d time.Duration) error {
	start := time.Now()
	waitcmd, _ := Command(List, Options{}, "")

	i := 0
	for {
		if err := waitcmd.Run(); err == nil {
			// no error, it's up!
			return nil
		}
		if time.Since(start) > d {
			return fmt.Errorf("no running systemd found within %s", d)
		}
		i++
		log.Printf("Loop %d, waiting for systemd to come up", i)
		time.Sleep(1 * time.Second)
	}
}
