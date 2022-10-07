package node

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/logging"
	"github.com/redesblock/mop/core/statestore/leveldb"
	"github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/storage"
)

// InitStateStore will initialize the stateStore with the given path to the
// data directory. When given an empty directory path, the function will instead
// initialize an in-memory state store that will not be persisted.
func InitStateStore(log logging.Logger, dataDir string) (ret storage.StateStorer, err error) {
	if dataDir == "" {
		ret = mock.NewStateStore()
		log.Warning("using in-mem state store, no node state will be persisted")
		return ret, nil
	}
	return leveldb.NewStateStore(filepath.Join(dataDir, "statestore"), log)
}

const overlayKey = "overlay"
const secureOverlayKey = "non-mineable-overlay"

// CheckOverlayWithStore checks the overlay is the same as stored in the statestore
func CheckOverlayWithStore(overlay flock.Address, storer storage.StateStorer) error {

	// migrate overlay key to new key
	err := storer.Delete(overlayKey)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	var storedOverlay flock.Address
	err = storer.Get(secureOverlayKey, &storedOverlay)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		return storer.Put(secureOverlayKey, overlay)
	}

	if !storedOverlay.Equal(overlay) {
		return fmt.Errorf("overlay address changed. was %s before but now is %s", storedOverlay, overlay)
	}

	return nil
}
