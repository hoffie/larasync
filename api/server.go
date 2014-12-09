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
	adminPubkey   [PubkeySize]byte
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
func New(adminPubkey [PubkeySize]byte, maxRequestAge time.Duration, rm *repository.Manager) *Server {
	serveMux := http.NewServeMux()
	s := Server{
		adminPubkey:   adminPubkey,
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
		if !ValidateRequest(req, s.adminPubkey, s.maxRequestAge) {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(rw, req)
	}
}

// ListenAndServe starts serving requests on the default port.
func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}
