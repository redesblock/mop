package manifest

import (
	"errors"
	"net/http"

	"github.com/redesblock/hop/core/swarm"
)

var ErrNotFound = errors.New("manifest: not found")

// Parser for manifest
type Parser interface {
	// Parse parses the encoded manifest data and returns the result
	Parse(bytes []byte) (Interface, error)
}

// Interface for operations with manifest
type Interface interface {
	// Add a manifest entry to specified path
	Add(string, Entry)
	// Remove reference from file on specified path
	Remove(string)
	// FindEntry returns manifest entry if one is found on specified path
	FindEntry(string) (Entry, error)
	// Serialize return encoded manifest
	Serialize() ([]byte, error)
}

// Entry represents single manifest entry
type Entry interface {
	// GetReference returns address of the entry file
	GetReference() swarm.Address
	// GetName returns the name of the file for the entry, if added
	GetName() string
	// GetHeaders returns the headers for manifest entry, if configured
	GetHeaders() http.Header
}
