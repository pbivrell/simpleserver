// server is a package that has convient methods for configuring and running a simple http/https server
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"
)

// Server is the struct for configuring a server
type Server struct {
	Port    int
	Addr    string
	Handler http.Handler
	// ShutdownTimeout timeout for server to gracefully shutdown
	ShutdownTimeout time.Duration
	srv             http.Server
}

type serverOpt func(s *Server)

// WithShutdwonTimeout configures Server with provided timeout
func WithShutdownTimeout(t time.Duration) serverOpt {
	return func(s *Server) {
		s.ShutdownTimeout = t
	}
}

// WithPort configures Server port
func WithPort(port int) serverOpt {
	return func(s *Server) {
		s.Port = port
	}
}

// WithAddr configures Server Addr
func WithAddr(addr string) serverOpt {
	return func(s *Server) {
		s.Addr = addr
	}
}

// WithHandler configures http.Handler
func WithHandler(h http.Handler) serverOpt {
	return func(s *Server) {
		s.Handler = h
	}
}

// WithCorsHandler wraps an http.Handler the configured cors options
func WithCorsHandler(h http.Handler, c cors.Options) serverOpt {
	return func(s *Server) {
		s.Handler = cors.New(c).Handler(h)
	}
}

// NewServer returns a server configured with sensible defaults. These defaults can be overriden with zero or more serverOpts
func NewServer(opts ...serverOpt) *Server {

	s := &Server{
		Port:            8080,
		Addr:            "",
		ShutdownTimeout: 15 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Shutdown allows the stopping a running server. It will attempt to gracefully shutdown with the configured ShutdownTimeout. Calling Shutdown before Run() will panic
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// WithSigShutdown will shutdown the running server when the provided sig happens. This call is blocking, so it is likely you will want to run it in a go routine in concert with server.Run()
func (s *Server) WithSigShutdown(sig os.Signal) error {

	c := make(chan os.Signal, 1)
	signal.Notify(c, sig)

	<-c

	return s.Shutdown()
}

// WithContextShutdown will shutdown the running server when the context is cancelled. This call is blocking, so it is likely you will want to run it in a go routine in concert with server.Run()
func (s *Server) WithContextShutdown(ctx context.Context) error {

	<-ctx.Done()

	return s.Shutdown()
}

// Run will start the http server and block until it has been Shutdown.
func (s *Server) Run() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Addr, s.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.Handler,
	}

	return srv.ListenAndServe()
}