package main

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	apicommon "github.com/hoffie/larasync/api/common"
	"github.com/hoffie/larasync/api/server"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"
)

// TestServer is used for creating and managing
// api.server.Server instances for testing.
type TestServer struct {
	listener    net.Listener
	hostAndPort string
	adminSecret []byte
	basePath    string
	rm          *repository.Manager
	api         *server.Server
}

// NewTestServer creates a server instance for testing purposes.
// It uses a random port for that.
func NewTestServer() (*TestServer, error) {
	ts := &TestServer{
		adminSecret: []byte("test secret"),
	}
	tempdir, err := ioutil.TempDir("", "lara")
	if err != nil {
		return nil, err
	}
	ts.basePath = tempdir

	rm, err := repository.NewManager(ts.basePath)
	if err != nil {
		return nil, err
	}

	pubKey, err := apicommon.GetAdminSecretPubkey(ts.adminSecret)
	if err != nil {
		return nil, err
	}

	certFile := filepath.Join(tempdir, "server.crt")
	keyFile := filepath.Join(tempdir, "server.key")
	err = x509.GenerateServerCertFiles(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	ts.api, err = server.New(pubKey, 5*time.Second, rm, certFile, keyFile)
	if err != nil {
		return nil, err
	}

	err = ts.makeListener()
	if err != nil {
		return nil, err
	}
	go ts.serve()
	return ts, nil
}

// serve starts serving requests on the listener.
func (ts *TestServer) serve() error {
	return ts.api.Serve(ts.listener)
}

// makeListener creates a new listener on a random port
// and saves the required address in the attribute hostAndPort
func (ts *TestServer) makeListener() error {
	// passing port :0 to Listen lets it choose a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	ts.listener = listener
	ts.hostAndPort = listener.Addr().String()
	return nil
}

// Close cleans when done using this instance;
// it removes the temporary directory and stops listening
func (ts *TestServer) Close() {
	ts.removeBasePath()
	ts.listener.Close()
}

// removeBasePath removes the temporary directory
func (ts *TestServer) removeBasePath() error {
	return os.RemoveAll(ts.basePath)
}
