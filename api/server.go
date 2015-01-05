package api

import (
	"fmt"
	"net/http"
	"os"
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

func jsonHeader(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
}

// setupRoutes is responsible for registering API endpoints.
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/repositories",
		s.requireAdminAuth(s.repositoryList)).Methods("GET")
	s.router.HandleFunc("/repositories/{repository}",
		s.requireAdminAuth(s.repositoryCreate)).Methods("PUT")

	s.router.HandleFunc("/repositories/{repository}/blobs/{blobID}",
		s.requireRepositoryAuth(s.blobGet)).Methods("GET")
	s.router.HandleFunc("/repositories/{repository}/blobs/{blobID}",
		s.requireRepositoryAuth(s.blobPut)).Methods("PUT")

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

// requireAuth wraps a handlerFunc and only calls it if the request is
// authenticated
func (s *Server) requireRepositoryAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		repositoryName := vars["repository"]
		repository, err := s.rm.Open(repositoryName)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(rw, "Repository not found", http.StatusNotFound)
			} else {
				http.Error(rw, "Internal Error", http.StatusInternalServerError)
			}
			return
		}

		var pubKeyArray [PubkeySize]byte
		pubKey, err := repository.GetAuthPubkey()
		if err != nil {
			http.Error(rw, "Internal Error", http.StatusInternalServerError)
		}

		copy(pubKeyArray[0:PubkeySize], pubKey[:PubkeySize])
		// TODO: Find if there is a better way for this.

		if !ValidateRequest(req, pubKeyArray, s.maxRequestAge) {
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
