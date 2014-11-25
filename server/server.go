package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/larasync/lara/sign/request"
)

// Server represents our http environment.
type Server struct {
	adminSecret   []byte
	router        *mux.Router
	maxRequestAge time.Duration
}

const (
	// DefaultPort specifies the server's default TCP port
	DefaultPort = 14124
)

// New returns a new Server.
func New(adminSecret []byte) *Server {
	s := Server{
		adminSecret:   adminSecret,
		maxRequestAge: time.Minute, //FIXME make configurable
		router:        mux.NewRouter(),
	}
	s.setupRoutes()
	http.Handle("/", s.router)
	return &s
}

// setupRoutes is responsible for registering API endpoints.
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/repositories",
		s.requireAdminAuth(s.repositoryList)).Methods("GET")
	s.router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("larasync\n"))
	})
}

// requireAdminAuth wraps a HandlerFunc and only calls it if the request has a
// valid admin auth header
func (s *Server) requireAdminAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if !request.ValidateAdminSigned(req, s.adminSecret, s.maxRequestAge) {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
}

// repositoryList returns a list of all configured repositories.
func (s *Server) repositoryList(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Not implemented") //FIXME implement
}

// ListenAndServe starts serving requests on the default port.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", DefaultPort), nil)
}
