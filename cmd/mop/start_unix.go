//go:build !windows
// +build !windows

package main

import (
	"errors"

	"github.com/redesblock/mop/core/logging"
)

func isWindowsService() (bool, error) {
	return false, nil
}

func createWindowsEventLogger(svcName string, logger logging.Logger) (logging.Logger, error) {
	return nil, errors.New("cannot create Windows event logger")
}
