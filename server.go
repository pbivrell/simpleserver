// simpleserver is a package that has convient methods for configuring and running a simple http/https server
package simpleserver

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
	Port         int
	Addr         string
	Handler      http.Handler
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	CertFile     string
	KeyFile      string
	// ShutdownTimeout timeout for server to gracefully shutdown
	ShutdownTimeout time.Duration
	srv             http.Server
}

type serverOpt func(s *Server)

// WithTLS configures server to use TLS
func WithTLS(certFile, keyFile string) serverOpt {
	return func(s *Server) {
		s.CertFile = certFile
		s.KeyFile = keyFile
	}
}

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

// WithWriteTimeout configures http.WriteTimeout
func WithWriteTimeout(t time.Duration) serverOpt {
	return func(s *Server) {
		s.WriteTimeout = t
	}
}

// WithReadTimeout configures http.WriteTimeout
func WithReadTimeout(t time.Duration) serverOpt {
	return func(s *Server) {
		s.ReadTimeout = t
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
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
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
		WriteTimeout: s.WriteTimeout,
		ReadTimeout:  s.ReadTimeout,
		IdleTimeout:  time.Second * 60,
		Handler:      s.Handler,
	}

	if s.CertFile != "" && s.KeyFile != "" {
		return srv.ListenAndServeTLS(s.CertFile, s.KeyFile)
	}

	return srv.ListenAndServe()
}
