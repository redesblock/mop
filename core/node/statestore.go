package node

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/storer/statestore/leveldb"
	"github.com/redesblock/mop/core/storer/storage"
)

// InitStateStore will initialize the stateStore with the given path to the
// data directory. When given an empty directory path, the function will instead
// initialize an in-memory state store that will not be persisted.
func InitStateStore(logger log.Logger, dataDir string) (storage.StateStorer, error) {
	if dataDir == "" {
		logger.Warning("using in-mem state store, no node state will be persisted")
		return leveldb.NewInMemoryStateStore(logger)
	}
	return leveldb.NewStateStore(filepath.Join(dataDir, "statestore"), logger)
}

const secureOverlayKey = "non-mineable-overlay"
const noncedOverlayKey = "nonce-overlay"

func GetExistingOverlay(storer storage.StateStorer) (cluster.Address, error) {
	var storedOverlay cluster.Address
	err := storer.Get(secureOverlayKey, &storedOverlay)
	if err != nil {
		return cluster.ZeroAddress, err
	}

	return storedOverlay, nil
}

// CheckOverlayWithStore checks the overlay is the same as stored in the statestore
func CheckOverlayWithStore(overlay cluster.Address, storer storage.StateStorer) error {

	var storedOverlay cluster.Address
	err := storer.Get(noncedOverlayKey, &storedOverlay)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		return storer.Put(noncedOverlayKey, overlay)
	}

	if !storedOverlay.Equal(overlay) {
		return fmt.Errorf("overlay address changed. was %s before but now is %s", storedOverlay, overlay)
	}

	return nil
}

// SetOverlayInStore sets the overlay stored in the statestore (for purpose of overlay migration)
func SetOverlayInStore(overlay cluster.Address, storer storage.StateStorer) error {
	return storer.Put(noncedOverlayKey, overlay)
}

const OverlayNonce = "overlayV2_nonce"

func overlayNonceExists(s storage.StateStorer) ([]byte, bool, error) {
	overlayNonce := make([]byte, 32)
	if err := s.Get(OverlayNonce, &overlayNonce); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return overlayNonce, true, nil
}

func setOverlayNonce(s storage.StateStorer, overlayNonce []byte) error {
	return s.Put(OverlayNonce, overlayNonce)
}
