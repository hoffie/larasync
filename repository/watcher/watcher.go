package watcher

import (
	"errors"
	"os"

	"github.com/cbrand/fsmonitor"
	"github.com/hoffie/larasync/constants"
	"gopkg.in/fsnotify.v1"
)

const (
	laraManagementDir = constants.LaraManagementDirName
)

// New returns a new watcher initialized with the passed
// configuration data.
func New(directoryPath string, handler RepositoryHandler) (*Watcher, error) {
	fsWatcher, err := fsmonitor.NewWatcherWithSkipFolders([]string{laraManagementDir})

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
		directoryPath:     directoryPath,
		handler:           handler,
		fileSystemWatcher: fsWatcher,
		Errors:            make(chan error),
		Close:             make(chan struct{}),
	}

	return watcher, nil
}

// Watcher is used to monitor item changes in larasync repositories
// and automatically keeps data in sync with the local internal repository
// state.
type Watcher struct {
	directoryPath     string
	handler           RepositoryHandler
	fileSystemWatcher *fsmonitor.Watcher
	// Errors which were emitted during processing the different
	// file system events.
	Errors chan error
	Close  chan struct{}
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
		case <-w.Close:
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
		close(w.Close)
	}
	return w.fileSystemWatcher.Close()
}
