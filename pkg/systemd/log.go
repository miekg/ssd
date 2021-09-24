package provider

import (
	"io"
	"log"
	"net/http"
	"time"
)

func (p *p) GetContainerLogsHandler(w http.ResponseWriter, r *http.Request) {
	handleError(func(w http.ResponseWriter, req *http.Request) error {
		r.Header.Set("Transfer-Encoding", "chunked")
		logsReader, cancel, err := journalReader(namespace, pod, container, opts)
		if err != nil {
			return err
		}
		defer logsReader.Close()
		defer cancel()

		// ResponseWriter must be flushed after each write.
		if _, ok := w.(writeFlusher); !ok {
			log.Printf("HTTP response writer does not support flushes")
		}
		fw := flushOnWrite(w)

		if !opts.Follow {
			io.Copy(fw, logsReader)
			return nil
		}

		// If in follow mode, follow until interrupted.
		untilTime := make(chan time.Time, 1)
		errChan := make(chan error, 1)

		go func(w io.Writer, errChan chan error) {
			err := journalFollow(untilTime, logsReader, w)
			errChan <- err
		}(fw, errChan)

		// Stop following logs if request context is completed.
		select {
		case err := <-errChan:
			return err
		case <-r.Context().Done():
			close(untilTime)
		}
		return nil
	})(w, r)
}

func flushOnWrite(w io.Writer) io.Writer {
	if fw, ok := w.(writeFlusher); ok {
		return &flushWriter{fw}
	}
	return w
}

type flushWriter struct{ w writeFlusher }

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
