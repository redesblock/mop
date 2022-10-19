//go:build !windows

package cmd

import (
	"errors"

	"github.com/redesblock/mop/core/log"
)

func isWindowsService() (bool, error) {
	return false, nil
}

func createWindowsEventLogger(_ string, _ log.Logger) (log.Logger, error) {
	return nil, errors.New("cannot create Windows event logger")
}
