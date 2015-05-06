package watcher

import (
	"errors"
	"os"

	"gopkg.in/fsnotify.v1"
	"github.com/cbrand/fsmonitor"
)

const (
	LARA_MANAGEMENT_DIR = ".lara"
)

// New returns a new watcher initialized with the passed
// configuration data.
func New(directoryPath string, handler RepositoryHandler) (*Watcher, error) {
	fsWatcher, err := fsmonitor.NewWatcherWithSkipFolders([]string{LARA_MANAGEMENT_DIR})

	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(directoryPath)

	if os.IsNotExist(err) {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, errors.New("The given path is not a directory")
	}

	watcher := &Watcher{
		directoryPath: directoryPath,
		handler: handler,
		fileSystemWatcher: fsWatcher,
		Errors: make(chan error),
		close: make(chan struct{}),
	}

	return watcher
}

// Watcher is used to monitor item changes in larasync repositories
// and automatically keeps data in sync with the local internal repository
// state.
type Watcher struct {
	directoryPath string
	handler RepositoryHandler
	fileSystemWatcher fsmonitor.Watcher
	// Errors which were emitted during processing the different
	// file system events.
	Errors chan error
	close chan struct{}
}

// Start initializes the internal filesystem watcher.
func (w *Watcher) Start() error {

	err := w.fileSystemWatcher.Watch(w.directoryPath)
	if err != nil {
		return err
	}

	go w.startLoop()

	return nil
}

// startLoop checks for the event messages and starts the management of incoming file system
// changes.
func (w *Watcher) startLoop() {
	for {
		var err error
		select {
			case ev := <-w.fileSystemWatcher.Events:
				if ev.Op == fsnotify.Write {
					err = w.handler.AddItem(ev.Name)
				} else if ev.Op == fsnotify.Remove {
					err = w.handler.DeleteItem(ev.Name)
				}
			case err = <-w.fileSystemWatcher.Error:
				// Handling done outside of select
			case <-w.close:
				return
		}

		if err != nil {
			w.Errors <- err
		}
	}
}

// Stop shuts down the internal processing loop.
func (w *Watcher) Stop() error {
	if !w.fileSystemWatcher.IsClosed() {
		close(w.close)
	}
	return w.fileSystemWatcher.Close()
}
