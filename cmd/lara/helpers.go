package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/repository"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	// PrivateKeySize is the size of the key used for signing.
	PrivateKeySize = repository.PrivateKeySize
	// PublicKeySize ist the size of the key used to verify the signature.
	PublicKeySize = repository.PublicKeySize
	// EncryptionKeySize is the key size used for encryption purposes.
	EncryptionKeySize = repository.EncryptionKeySize
)

// clientFor returns the Client which is configured to communicate
// with the given server repository.
func (d *Dispatcher) clientFor(r *repository.ClientRepository) (*api.Client, error) {
	sc, err := r.StateConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load state config (%s)", err)
	}
	if sc.DefaultServer == "" {
		return nil, fmt.Errorf("no default server configured (state)")
	}
	privKey, err := r.GetSigningPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get signing private key (%s)", err)
	}
	client := d.clientForState(sc)
	client.SetSigningPrivateKey(privKey)

	return client, nil
}

func (d *Dispatcher) clientForState(sc *repository.StateConfig) *api.Client {
	d.sc = sc
	return api.NewClient(sc.DefaultServer, sc.DefaultServerFingerprint,
		d.confirmFingerprint)
}

// promptPassword outputs the given prompt text and waits for a value to be entered
// on the input stream. It attempts to do so securely.
func (d *Dispatcher) promptPassword(prompt string) ([]byte, error) {
	switch d.stdin.(type) {
	case *os.File:
		return d.promptGetpass(prompt)
	}
	return d.promptCleartext(prompt)
}

// promptGetpass reads a password from our input,
// attempting to hide the input if possible.
func (d *Dispatcher) promptGetpass(prompt string) ([]byte, error) {
	file := d.stdin.(*os.File)
	fd := int(file.Fd())
	if !terminal.IsTerminal(fd) {
		return d.promptCleartext(prompt)
	}
	d.stdout.Write([]byte(prompt))
	defer d.stdout.Write([]byte("\n"))
	return terminal.ReadPassword(fd)
}

// promptCleartext reads text from our input in the standard way.
func (d *Dispatcher) promptCleartext(prompt string) ([]byte, error) {
	d.stdout.Write([]byte(prompt))
	line, err := d.readLine()
	if err != nil {
		return nil, err
	}
	return line[:len(line)], nil
}

// readLine reads exactly one line; it does not return the delimiter.
// The difference to other methods for similar goals (bufio.Scanner or ReadBytes)
// is that it only reads as much data as is needed, i.e. it keeps any left-over
// data in the original reader for later access.
func (d *Dispatcher) readLine() ([]byte, error) {
	buf := []byte{}
	oneChar := make([]byte, 1)
	for {
		_, err := d.stdin.Read(oneChar)
		if err != nil {
			return nil, err
		}
		if oneChar[0] == '\n' {
			return buf, nil
		}
		buf = append(buf, oneChar[0])
	}
}

// confirmFingerprint is called to ask for permission to connect to a new server.
func (d *Dispatcher) confirmFingerprint(fp string) bool {
	fmt.Fprintf(d.stdout, "You are connecting to this server for the first time.\n")
	fmt.Fprintf(d.stdout, "To avoid any attacks on the connection between you and the server,\n")
	fmt.Fprint(d.stdout, "please verify the fingerprint.\n")
	fmt.Fprint(d.stdout, "Do this by getting the fingerprint on the server using\n")
	fmt.Fprint(d.stdout, "  lara server-fingerprint\n")
	fmt.Fprint(d.stdout, "or ask the server administrator to provide you with the fingerprint.\n\n")
	fmt.Fprint(d.stdout, "Only continue connecting when the above mentioned fingerprint matches this value:\n")
	fmt.Fprintf(d.stdout, "  %s\n", fp)
	res, err := d.promptCleartext("Do the fingerprints match? (y/N) ")
	if err != nil {
		return false
	}
	accept := reflect.DeepEqual(res, []byte("y"))
	if !accept {
		return false
	}
	if d.sc == nil {
		fmt.Fprintf(d.stderr,
			"Warning: cannot save fingerprint (no state config)\n")
		// yes, we return true here nevertheless
		return true
	}
	d.sc.DefaultServerFingerprint = fp
	err = d.sc.Save()
	if err != nil {
		fmt.Fprintf(d.stderr, "Warning: failed to save fingerprint (%s)\n", err)
	}
	return true
}
