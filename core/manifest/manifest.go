package manifest

import (
	"encoding"
	"errors"
	"net/http"

	"github.com/redesblock/hop/core/swarm"
)

// ErrNotFound is returned when an Entry is not found in the manifest.
var ErrNotFound = errors.New("manifest: not found")

// Interface for operations with manifest.
type Interface interface {
	// Add a manifest entry to the specified path.
	Add(string, Entry)
	// Remove a manifest entry on the specified path.
	Remove(string)
	// Entry returns a manifest entry if one is found in the specified path.
	Entry(string) (Entry, error)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// Entry represents a single manifest entry.
type Entry interface {
	// Reference returns the address of the file in the entry.
	Reference() swarm.Address
	// Name returns the name of the file in the entry.
	Name() string
	// Headers returns the headers for the file in the manifest entry.
	Headers() http.Header
}
