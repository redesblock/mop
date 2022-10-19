package multiresolver

import "github.com/redesblock/mop/core/log"

func GetLogger(mr *MultiResolver) log.Logger {
	return mr.logger
}

func GetCfgs(mr *MultiResolver) []ConnectionConfig {
	return mr.cfgs
}
