package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func serveUntilSignal(servers []*http.Server, shutdownTimeout time.Duration, backgroundErrCh ...<-chan error) error {
	if len(servers) == 0 {
		return errors.New("riido_ai_server: at least one http server is required")
	}
	bgErrCh := mergeBackgroundErrors(backgroundErrCh...)
	errCh := listenAndServeAll(servers)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		if err := shutdownServers(servers, shutdownTimeout); err != nil {
			return fmt.Errorf("shutdown after %s: %w", sig, err)
		}
		return waitForServers(errCh, len(servers))
	case err := <-bgErrCh:
		return shutdownAfterBackgroundError(servers, shutdownTimeout, errCh, err)
	case err := <-errCh:
		_ = shutdownServers(servers, shutdownTimeout)
		return err
	}
}

func listenAndServeAll(servers []*http.Server) <-chan error {
	errCh := make(chan error, len(servers))
	for _, server := range servers {
		go func(server *http.Server) {
			var err error
			if server.TLSConfig == nil {
				err = server.ListenAndServe()
			} else {
				err = server.ListenAndServeTLS("", "")
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
				return
			}
			errCh <- nil
		}(server)
	}
	return errCh
}
