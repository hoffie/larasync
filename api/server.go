package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/hoffie/larasync/helpers/lock"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"
)

const (
	// DefaultPort specifies the server's default TCP port
	DefaultPort = 14124
)

// Server represents our http environment.
type Server struct {
	adminPubkey   [PublicKeySize]byte
	router        *mux.Router
	maxRequestAge time.Duration
	http          *http.Server
	certFile      string
	keyFile       string
	certificate   tls.Certificate
	rm            *repository.Manager
}

// New returns a new Server.
func New(adminPubkey [PublicKeySize]byte, maxRequestAge time.Duration, rm *repository.Manager, certFile, keyFile string) (*Server, error) {
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
		certFile: certFile,
		keyFile:  keyFile,
	}
	s.setupRoutes()
	err := s.loadCertificate()
	if err != nil {
		return nil, err
	}
	serveMux.Handle("/", s.router)
	return &s, nil
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

	s.router.HandleFunc("/repositories/{repository}/nibs",
		s.requireRepositoryAuth(s.nibList)).Methods("GET")
	s.router.HandleFunc("/repositories/{repository}/nibs/{nibID}",
		s.requireRepositoryAuth(s.nibGet)).Methods("GET")
	s.router.HandleFunc("/repositories/{repository}/nibs/{nibID}",
		s.requireRepositoryAuth(
			s.synchronizeWith(
				"nibPUT",
				s.checkTransactionPrecondition(s.nibPut),
			),
		),
	).Methods("PUT")

	s.router.HandleFunc("/repositories/{repository}/authorizations/{authPublicKey}",
		s.authorizationGet).Methods("GET")
	s.router.HandleFunc("/repositories/{repository}/authorizations/{authPublicKey}",
		s.requireRepositoryAuth(s.authorizationPut)).Methods("PUT")

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
				// Repository is not found. However, due to security reasons
				// we are returning Unauthorized here so that an unauthenticated user
				// cannot check wether a repository does exist or not.
				// FIXME: timing side channel
				http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			} else {
				http.Error(rw, "Internal Error", http.StatusInternalServerError)
			}
			return
		}

		var pubKeyArray [PublicKeySize]byte
		pubKey, err := repository.GetSigningPublicKey()
		if err != nil {
			http.Error(rw, "Internal Error", http.StatusInternalServerError)
			return
		}

		copy(pubKeyArray[0:PublicKeySize], pubKey[:PublicKeySize])
		// TODO: Find if there is a better way for this.

		if !ValidateRequest(req, pubKeyArray, s.maxRequestAge) {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		f(rw, req)
	}
}

// synchronizeWith can be used as a wrapper. The endpoint will then be synchronized
// via the lock manager and the given role and in the given repository.
func (s *Server) synchronizeWith(roleName string, f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		repositoryName := vars["repository"]
		repository, err := s.rm.Open(repositoryName)
		if err != nil {
			http.Error(rw, "Internal Error", http.StatusInternalServerError)
			return
		}

		manager := lock.CurrentManager()
		lock := manager.Get(
			repository.GetManagementDir(),
			fmt.Sprintf("api:%s", roleName),
		)

		lock.Lock()
		f(rw, req)
		lock.Unlock()
	}
}

// checkTransactionPrecondition checks if there is a if-match header set that this
// entry is the current transaction id in the system.
func (s *Server) checkTransactionPrecondition(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		repositoryName := vars["repository"]
		repository, err := s.rm.Open(repositoryName)
		if err != nil {
			http.Error(rw, "Internal Error", http.StatusInternalServerError)
			return
		}

		header := req.Header
		ifMatch := header.Get("If-Match")
		if ifMatch != "" {
			currentTransaction, err := repository.CurrentTransaction()
			if err != nil {
				http.Error(rw, "Internal Error", http.StatusInternalServerError)
				return
			}

			if ifMatch != currentTransaction.IDString() {
				rw.WriteHeader(http.StatusPreconditionFailed)
				return
			}
		}

		f(rw, req)
	}
}

// attachCurrentTransactionHeader is used to notify
func attachCurrentTransactionHeader(r *repository.Repository, rw http.ResponseWriter) {
	header := rw.Header()

	currentTransaction, err := r.CurrentTransaction()
	if err != nil && err != repository.ErrTransactionNotExists {
		errorText(rw, "Internal Error", http.StatusInternalServerError)
		return
	} else if currentTransaction != nil {
		header.Set("X-Current-Transaction-Id", currentTransaction.IDString())
	}
}

// ListenAndServe starts serving requests on the default port using TLS
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// loadCertificate loads and parses the on-disk certificates and keeps them
// in memory for later use.
func (s *Server) loadCertificate() error {
	var err error
	s.certificate, err = tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return err
	}
	if len(s.certificate.Certificate) != 1 {
		return fmt.Errorf("bad number of certificates loaded (%d)",
			len(s.certificate.Certificate))
	}
	fp, err := s.CertificateFingerprint()
	Log.Info("loaded certificate", log15.Ctx{"fingerprint": fp})
	return nil
}

// CertificateFingerprint returns the server's certificate public key fingerprint
func (s *Server) CertificateFingerprint() (string, error) {
	fp, err := x509.CertificateFingerprintFromBytes(s.certificate.Certificate[0])
	if err != nil {
		return "", err
	}
	return fp, nil
}

// Serve serves TLS-enabled requests on the given listener.
func (s *Server) Serve(l net.Listener) error {
	config := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{s.certificate},
	}
	tlsListener := tls.NewListener(tcpKeepAliveListener{l.(*net.TCPListener)}, config)
	return s.http.Serve(tlsListener)
}
