package mock

import (
	"context"
	"time"

	"github.com/redesblock/mop/core/swarm"
)

type Service struct {
	pingFunc func(ctx context.Context, address swarm.Address, msgs ...string) (rtt time.Duration, err error)
}

func New(pingFunc func(ctx context.Context, address swarm.Address, msgs ...string) (rtt time.Duration, err error)) *Service {
	return &Service{pingFunc: pingFunc}
}

func (s *Service) Ping(ctx context.Context, address swarm.Address, msgs ...string) (rtt time.Duration, err error) {
	return s.pingFunc(ctx, address, msgs...)
}
