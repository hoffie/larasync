package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hoffie/larasync/repository"
)

// Server represents our http environment.
type Server struct {
	adminSecret   []byte
	router        *mux.Router
	maxRequestAge time.Duration
	http          *http.Server
	rm            *repository.Manager
}

const (
	// DefaultPort specifies the server's default TCP port
	DefaultPort = 14124
)

// New returns a new Server.
func New(adminSecret []byte, maxRequestAge time.Duration, rm *repository.Manager) *Server {
	serveMux := http.NewServeMux()
	s := Server{
		adminSecret:   adminSecret,
		maxRequestAge: maxRequestAge,
		rm:            rm,
		router:        mux.NewRouter(),
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", DefaultPort),
			Handler: serveMux,
		},
	}
	s.setupRoutes()
	serveMux.Handle("/", s.router)
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
		if !ValidateAdminSigned(req, s.adminSecret, s.maxRequestAge) {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(rw, req)
	}
}

// repositoryList returns a list of all configured repositories.
func (s *Server) repositoryList(rw http.ResponseWriter, req *http.Request) {
	names, err := s.rm.ListNames()
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(names)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	rw.Write(out)
}

// ListenAndServe starts serving requests on the default port.
func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}
