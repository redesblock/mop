package chainsyncer_test

import (
	"context"
	"errors"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/mop/core/chainsyncer"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/logging"
	"github.com/redesblock/mop/core/topology/mock"
	"github.com/redesblock/mop/core/transaction/backendmock"
)

func TestChainsyncer(t *testing.T) {
	var (
		expBlockHash    = common.HexToHash("0x9de2787d1d80a6164f4bc6359d9017131cbc14402ee0704bff0c6d691701c1dc").Bytes()
		logger          = logging.New(ioutil.Discard, 0)
		trxBlock        = common.HexToHash("0x2")
		blockC          = make(chan struct{})
		nextBlockHeader = &types.Header{ParentHash: trxBlock}
		blockNumber     = backendmock.WithBlockNumberFunc(func(_ context.Context) (uint64, error) { return uint64(100), nil })
		headerByNum     = backendmock.WithHeaderbyNumberFunc(func(ctx context.Context, number *big.Int) (*types.Header, error) {
			return nextBlockHeader, nil
		})

		backend        = backendmock.New(headerByNum, blockNumber)
		topology       = mock.NewTopologyDriver(mock.WithPeers(flock.NewAddress([]byte{0, 1, 2, 3})))
		proofBlockHash = make([]byte, 32)
		proofError     = errors.New("error")
		p              = &prover{f: func(_ flock.Address, _ uint64) ([]byte, error) {
			return proofBlockHash, proofError
		}}
		d = &m{f: func(_ flock.Address, _ time.Duration) {
			select {
			case blockC <- struct{}{}:
			default:
			}
		}}
	)

	newChainSyncerTest := func(e error, blockHash []byte, cb func(*testing.T)) func(*testing.T) {
		proofError = e
		proofBlockHash = blockHash
		return func(t *testing.T) {
			cs, err := chainsyncer.New(backend, p, topology, d, logger, &chainsyncer.Options{
				FlagTimeout:     500 * time.Millisecond,
				PollEvery:       100 * time.Millisecond,
				BlockerPollTime: 100 * time.Millisecond,
			})
			if err != nil {
				t.Fatal(err)
			}

			defer cs.Close()
			cb(t)
		}
	}

	t.Run("prover error", newChainSyncerTest(proofError, proofBlockHash, func(t *testing.T) {
		select {
		case <-blockC:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for blocklisting")
		}
	}))

	t.Run("blockhash mismatch", newChainSyncerTest(nil, proofBlockHash, func(t *testing.T) {
		select {
		case <-blockC:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for blocklisting")
		}
	}))

	t.Run("all good", newChainSyncerTest(nil, expBlockHash, func(t *testing.T) {
		select {
		case <-blockC:
			t.Fatal("blocklisting occurred but should not have")
		case <-time.After(500 * time.Millisecond):
		}
	}))
}

type prover struct {
	f func(flock.Address, uint64) ([]byte, error)
}

func (p *prover) Prove(_ context.Context, a flock.Address, b uint64) ([]byte, error) {
	return p.f(a, b)
}

type m struct {
	f func(flock.Address, time.Duration)
}

func (m *m) Disconnect(overlay flock.Address, reason string) error {
	panic("not implemented")
}
func (m *m) Blocklist(overlay flock.Address, duration time.Duration, reason string) error {
	m.f(overlay, duration)
	return nil
}
