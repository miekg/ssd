package systemd

import (
	"bufio"
	"errors"
	"io"
	"time"
)

// Follow synchronously follows the io.Reader, writing each new journal entry to writer. The
// follow will continue until a single time.Time is received on the until channel (or it's closed).
func Follow(until <-chan time.Time, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	bufch := make(chan []byte)
	errch := make(chan error)

	go func() {
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				errch <- err
				return
			}
			bufch <- scanner.Bytes()
		}
		// When the context is Done() the 'until' channel is closed, this kicks in the defers in the GetContainerLogsHandler method.
		// this cleans up the journalctl, and closes all file descripters. Scan() then stops with an error (before any reads,
		// hence the above if err .. .isn't triggered). In the end this go-routine exits.
		// the error here is "read |0: file already closed".
	}()

	for {
		select {
		case <-until:
			return errors.New("timeout expired")

		case err := <-errch:
			return err

		case buf := <-bufch:
			if _, err := w.Write(buf); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}
}

func FlushWriter(w io.Writer) io.Writer {
	if fw, ok := w.(writeFlusher); ok {
		return &flushWriter{fw}
	}
	return w
}

type flushWriter struct {
	w writeFlusher
}

type writeFlusher interface {
	Flush()
	Write([]byte) (int, error)
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	if n > 0 {
		fw.w.Flush()
	}
	return n, err
}
