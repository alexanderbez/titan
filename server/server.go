package server

import (
	"context"
	"net/http"
	"time"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
)

// Server is a simple wrapper around an embedded HTTP server with a logger and
// database.
type Server struct {
	*http.Server
	db     core.DB
	logger core.Logger
}

// CreateServer attempts to start a RESTful JSON HTTP service. If the server
// fails to start, an error is returned.
func CreateServer(cfg config.Config, db core.DB, logger core.Logger) (*Server, error) {
	srvr := &Server{
		db:     db,
		logger: logger.With("module", "server"),
	}

	srvr.Server = &http.Server{
		Addr:         cfg.Network.ListenAddr,
		Handler:      srvr.createRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	errCh := make(chan error)

	// start the server in a go-routine and write to a channel if an error occurs
	go func() {
		if err := srvr.Server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	// Check if there was an error returned by starting the server or wait enough
	// time to safely return.
	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(1 * time.Second):
		return srvr, nil
	}
}

// Close attempts to perform a clean shutdown of the server.
func (srvr *Server) Close() {
	srvr.logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srvr.SetKeepAlivesEnabled(false)
	if err := srvr.Shutdown(ctx); err != nil {
		srvr.logger.Fatalf("failed to gracefully shutdown server: %v\n", err)
	}
}
