package node

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-multierror"
	"github.com/multiformats/go-multiaddr"
	mop "github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/auth"
	"github.com/redesblock/mop/core/chain/transaction"
	"github.com/redesblock/mop/core/chain/transaction/backendmock"
	transactionmock "github.com/redesblock/mop/core/chain/transaction/mock"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds/factory"
	mockAccounting "github.com/redesblock/mop/core/incentives/bookkeeper/mock"
	"github.com/redesblock/mop/core/incentives/settlement/swap/chequebook"
	mockchequebook "github.com/redesblock/mop/core/incentives/settlement/swap/chequebook/mock"
	erc20mock "github.com/redesblock/mop/core/incentives/settlement/swap/erc20/mock"
	swapmock "github.com/redesblock/mop/core/incentives/settlement/swap/mock"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/incentives/voucher/batchstore"
	mockPost "github.com/redesblock/mop/core/incentives/voucher/mock"
	vouchertesting "github.com/redesblock/mop/core/incentives/voucher/testing"
	"github.com/redesblock/mop/core/incentives/voucher/vouchercontract"
	mockPostContract "github.com/redesblock/mop/core/incentives/voucher/vouchercontract/mock"
	"github.com/redesblock/mop/core/log"
	mockP2P "github.com/redesblock/mop/core/p2p/mock"
	"github.com/redesblock/mop/core/p2p/topology/lightnode"
	mockTopology "github.com/redesblock/mop/core/p2p/topology/mock"
	pinning "github.com/redesblock/mop/core/pins/mock"
	mockPingPong "github.com/redesblock/mop/core/protocol/pingpong/mock"
	"github.com/redesblock/mop/core/protocol/pseudosettle"
	"github.com/redesblock/mop/core/protocol/pushsync"
	mockPushsync "github.com/redesblock/mop/core/protocol/pushsync/mock"
	"github.com/redesblock/mop/core/psser"
	resolverMock "github.com/redesblock/mop/core/resolver/mock"
	"github.com/redesblock/mop/core/storer/localstore"
	"github.com/redesblock/mop/core/storer/statestore/leveldb"
	mockStateStore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/tracer"
	"github.com/redesblock/mop/core/traverser"
	"github.com/redesblock/mop/core/util/ioutil"
	mockSteward "github.com/redesblock/mop/core/warden/mock"
	"golang.org/x/sync/errgroup"
)

type DevMop struct {
	tracerCloser     io.Closer
	stateStoreCloser io.Closer
	localstoreCloser io.Closer
	apiCloser        io.Closer
	pssCloser        io.Closer
	tagsCloser       io.Closer
	errorLogWriter   io.Writer
	apiServer        *http.Server
	debugAPIServer   *http.Server
}

type DevOptions struct {
	Logger                   log.Logger
	APIAddr                  string
	DebugAPIAddr             string
	CORSAllowedOrigins       []string
	DBOpenFilesLimit         uint64
	ReserveCapacity          uint64
	DBWriteBufferSize        uint64
	DBBlockCacheCapacity     uint64
	DBDisableSeeksCompaction bool
	Restricted               bool
	TokenEncryptionKey       string
	AdminPasswordHash        string
}

// NewDevMop starts the mop instance in 'development' mode
// this implies starting an API and a Debug endpoints while mocking all their services.
func NewDevMop(logger log.Logger, o *DevOptions) (b *DevMop, err error) {
	tracer, tracerCloser, err := tracer.NewTracer(&tracer.Options{
		Enabled: false,
	})
	if err != nil {
		return nil, fmt.Errorf("tracer: %w", err)
	}

	sink := ioutil.WriterFunc(func(p []byte) (int, error) {
		logger.Error(nil, string(p))
		return len(p), nil
	})

	b = &DevMop{
		errorLogWriter: sink,
		tracerCloser:   tracerCloser,
	}

	stateStore, err := leveldb.NewInMemoryStateStore(logger)
	if err != nil {
		return nil, err
	}
	b.stateStoreCloser = stateStore

	batchStore, err := batchstore.New(stateStore, func(b []byte) error { return nil }, logger)
	if err != nil {
		return nil, fmt.Errorf("batchstore: %w", err)
	}

	err = batchStore.PutChainState(&voucher.ChainState{
		CurrentPrice: big.NewInt(1),
		TotalAmount:  big.NewInt(1),
	})
	if err != nil {
		return nil, fmt.Errorf("batchstore: %w", err)
	}

	mockKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		return nil, err
	}
	signer := crypto.NewDefaultSigner(mockKey)

	overlayBSCAddress, err := signer.BSCAddress()
	if err != nil {
		return nil, fmt.Errorf("BNB Smart Chain address: %w", err)
	}

	var authenticator *auth.Authenticator

	if o.Restricted {
		if authenticator, err = auth.New(o.TokenEncryptionKey, o.AdminPasswordHash, logger); err != nil {
			return nil, fmt.Errorf("authenticator: %w", err)
		}
		logger.Info("starting with restricted APIs")
	}

	var mockTransaction = transactionmock.New(transactionmock.WithPendingTransactionsFunc(func() ([]common.Hash, error) {
		return []common.Hash{common.HexToHash("abcd")}, nil
	}), transactionmock.WithResendTransactionFunc(func(ctx context.Context, txHash common.Hash) error {
		return nil
	}), transactionmock.WithStoredTransactionFunc(func(txHash common.Hash) (*transaction.StoredTransaction, error) {
		recipient := common.HexToAddress("dfff")
		return &transaction.StoredTransaction{
			To:          &recipient,
			Created:     1,
			Data:        []byte{1, 2, 3, 4},
			GasPrice:    big.NewInt(12),
			GasLimit:    5345,
			Value:       big.NewInt(4),
			Nonce:       3,
			Description: "test",
		}, nil
	}), transactionmock.WithCancelTransactionFunc(func(ctx context.Context, originalTxHash common.Hash) (common.Hash, error) {
		return common.Hash{}, nil
	}),
	)

	chainBackend := backendmock.New(
		backendmock.WithBlockNumberFunc(func(ctx context.Context) (uint64, error) {
			return 1, nil
		}),
		backendmock.WithBalanceAt(func(ctx context.Context, address common.Address, block *big.Int) (*big.Int, error) {
			return big.NewInt(0), nil
		}),
	)

	probe := api.NewProbe()
	probe.SetHealthy(api.ProbeStatusOK)
	defer func(probe *api.Probe) {
		if err != nil {
			probe.SetHealthy(api.ProbeStatusNOK)
		} else {
			probe.SetReady(api.ProbeStatusOK)
		}
	}(probe)

	var debugApiService *api.Service

	if o.DebugAPIAddr != "" {
		debugAPIListener, err := net.Listen("tcp", o.DebugAPIAddr)
		if err != nil {
			return nil, fmt.Errorf("debug api listener: %w", err)
		}

		debugApiService = api.New(mockKey.PublicKey, mockKey.PublicKey, overlayBSCAddress, logger, mockTransaction, batchStore, api.DevMode, true, true, chainBackend, o.CORSAllowedOrigins)
		debugAPIServer := &http.Server{
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           debugApiService,
			ErrorLog:          stdlog.New(b.errorLogWriter, "", 0),
		}

		debugApiService.MountTechnicalDebug()
		debugApiService.SetProbe(probe)

		go func() {
			logger.Info("starting debug api server", "address", debugAPIListener.Addr())

			if err := debugAPIServer.Serve(debugAPIListener); err != nil && err != http.ErrServerClosed {
				logger.Debug("debug api server failed to start", "error", err)
				logger.Error(nil, "debug api server failed to start")
			}
		}()

		b.debugAPIServer = debugAPIServer
	}

	lo := &localstore.Options{
		Capacity:               1000000,
		ReserveCapacity:        o.ReserveCapacity,
		OpenFilesLimit:         o.DBOpenFilesLimit,
		BlockCacheCapacity:     o.DBBlockCacheCapacity,
		WriteBufferSize:        o.DBWriteBufferSize,
		DisableSeeksCompaction: o.DBDisableSeeksCompaction,
		UnreserveFunc: func(voucher.UnreserveIteratorFn) error {
			return nil
		},
	}

	var clusterAddress cluster.Address
	storer, err := localstore.New("", clusterAddress.Bytes(), stateStore, lo, logger)
	if err != nil {
		return nil, fmt.Errorf("localstore: %w", err)
	}
	b.localstoreCloser = storer

	tagService := tags.NewTags(stateStore, logger)
	b.tagsCloser = tagService

	pssService := psser.New(mockKey, logger)
	b.pssCloser = pssService

	pssService.SetPushSyncer(mockPushsync.New(func(ctx context.Context, chunk cluster.Chunk) (*pushsync.Receipt, error) {
		pssService.TryUnwrap(chunk)
		return &pushsync.Receipt{}, nil
	}))

	traversalService := traverser.New(storer)

	post := mockPost.New()
	voucherContract := mockPostContract.New(
		mockPostContract.WithCreateBatchFunc(
			func(ctx context.Context, amount *big.Int, depth uint8, immutable bool, label string) ([]byte, error) {
				id := vouchertesting.MustNewID()
				batch := &voucher.Batch{
					ID:        id,
					Owner:     overlayBSCAddress.Bytes(),
					Value:     big.NewInt(0).Mul(amount, big.NewInt(int64(1<<depth))),
					Depth:     depth,
					Immutable: immutable,
				}

				err := batchStore.Save(batch)
				if err != nil {
					return nil, err
				}

				stampIssuer := voucher.NewStampIssuer(label, string(overlayBSCAddress.Bytes()), id, amount, batch.Depth, 0, 0, immutable)
				_ = post.Add(stampIssuer)

				return id, nil
			},
		),
		mockPostContract.WithTopUpBatchFunc(
			func(ctx context.Context, batchID []byte, topupAmount *big.Int) error {
				batch, err := batchStore.Get(batchID)
				if err != nil {
					return err
				}

				totalAmount := big.NewInt(0).Mul(topupAmount, big.NewInt(int64(1<<batch.Depth)))

				newBalance := big.NewInt(0).Add(totalAmount, batch.Value)

				err = batchStore.Update(batch, newBalance, batch.Depth)
				if err != nil {
					return err
				}
				topUpAmount := big.NewInt(0).Div(batch.Value, big.NewInt(int64(1<<(batch.Depth))))

				post.HandleTopUp(batch.ID, topUpAmount)
				return nil
			},
		),
		mockPostContract.WithDiluteBatchFunc(
			func(ctx context.Context, batchID []byte, newDepth uint8) error {
				batch, err := batchStore.Get(batchID)
				if err != nil {
					return err
				}

				if newDepth < batch.Depth {
					return vouchercontract.ErrInvalidDepth
				}

				newBalance := big.NewInt(0).Div(batch.Value, big.NewInt(int64(1<<(newDepth-batch.Depth))))

				err = batchStore.Update(batch, newBalance, newDepth)
				if err != nil {
					return err
				}

				post.HandleDepthIncrease(batch.ID, newDepth)
				return nil
			},
		),
	)

	var (
		lightNodes = lightnode.NewContainer(cluster.NewAddress(nil))
		pingPong   = mockPingPong.New(pong)
		p2ps       = mockP2P.New(
			mockP2P.WithConnectFunc(func(ctx context.Context, addr multiaddr.Multiaddr) (address *mop.Address, err error) {
				return &mop.Address{}, nil
			}), mockP2P.WithDisconnectFunc(
				func(cluster.Address, string) error {
					return nil
				},
			), mockP2P.WithAddressesFunc(
				func() ([]multiaddr.Multiaddr, error) {
					ma, _ := multiaddr.NewMultiaddr("mock")
					return []multiaddr.Multiaddr{ma}, nil
				},
			))
		acc            = mockAccounting.NewAccounting()
		kad            = mockTopology.NewTopologyDriver()
		storeRecipient = mockStateStore.NewStateStore()
		pseudoset      = pseudosettle.New(nil, logger, storeRecipient, nil, big.NewInt(10000), big.NewInt(10000), p2ps)
		mockSwap       = swapmock.New(swapmock.WithCashoutStatusFunc(
			func(ctx context.Context, peer cluster.Address) (*chequebook.CashoutStatus, error) {
				return &chequebook.CashoutStatus{
					Last:           &chequebook.LastCashout{},
					UncashedAmount: big.NewInt(0),
				}, nil
			},
		), swapmock.WithLastSentChequeFunc(
			func(a cluster.Address) (*chequebook.SignedCheque, error) {
				return &chequebook.SignedCheque{
					Cheque: chequebook.Cheque{
						Beneficiary: common.Address{},
						Chequebook:  common.Address{},
					},
				}, nil
			},
		), swapmock.WithLastReceivedChequeFunc(
			func(a cluster.Address) (*chequebook.SignedCheque, error) {
				return &chequebook.SignedCheque{
					Cheque: chequebook.Cheque{
						Beneficiary: common.Address{},
						Chequebook:  common.Address{},
					},
				}, nil
			},
		))
		mockChequebook = mockchequebook.NewChequebook(mockchequebook.WithChequebookBalanceFunc(
			func(context.Context) (ret *big.Int, err error) {
				return big.NewInt(0), nil
			},
		), mockchequebook.WithChequebookAvailableBalanceFunc(
			func(context.Context) (ret *big.Int, err error) {
				return big.NewInt(0), nil
			},
		), mockchequebook.WithChequebookWithdrawFunc(
			func(ctx context.Context, amount *big.Int) (hash common.Hash, err error) {
				return common.Hash{}, nil
			},
		), mockchequebook.WithChequebookDepositFunc(
			func(ctx context.Context, amount *big.Int) (hash common.Hash, err error) {
				return common.Hash{}, nil
			},
		))
	)

	var (
		// syncStatusFn mocks chainsync status because complete chainsync is required in order to curl certain apis e.g. /stamps.
		// this allows accessing those apis by passing true to isDone in devNode.
		syncStatusFn = func() (isDone bool, err error) {
			return true, nil
		}
	)

	mockFeeds := factory.New(storer)
	mockResolver := resolverMock.NewResolver()
	mockPinning := pinning.NewServiceMock()
	mockSteward := new(mockSteward.Steward)

	debugOpts := api.ExtraOptions{
		Pingpong:         pingPong,
		TopologyDriver:   kad,
		LightNodes:       lightNodes,
		Accounting:       acc,
		Pseudosettle:     pseudoset,
		Swap:             mockSwap,
		Chequebook:       mockChequebook,
		BlockTime:        big.NewInt(2),
		Tags:             tagService,
		Storer:           storer,
		Resolver:         mockResolver,
		Pss:              pssService,
		TraversalService: traversalService,
		Pinning:          mockPinning,
		FeedFactory:      mockFeeds,
		Post:             post,
		VoucherContract:  voucherContract,
		Warden:           mockSteward,
		SyncStatus:       syncStatusFn,
	}

	var erc20 = erc20mock.New(
		erc20mock.WithBalanceOfFunc(func(ctx context.Context, address common.Address) (*big.Int, error) {
			return big.NewInt(0), nil
		}),
		erc20mock.WithTransferFunc(func(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error) {
			return common.Hash{}, nil
		}),
	)

	apiService := api.New(mockKey.PublicKey, mockKey.PublicKey, overlayBSCAddress, logger, mockTransaction, batchStore, api.DevMode, true, true, chainBackend, o.CORSAllowedOrigins)

	apiService.Configure(signer, authenticator, tracer, api.Options{
		CORSAllowedOrigins: o.CORSAllowedOrigins,
		WsPingPeriod:       60 * time.Second,
		Restricted:         o.Restricted,
	}, debugOpts, 1, erc20)

	apiService.MountAPI()
	apiService.SetProbe(probe)

	if o.Restricted {
		apiService.SetP2P(p2ps)
		apiService.SetClusterAddress(&clusterAddress)
		apiService.MountDebug(true)
	}

	if o.DebugAPIAddr != "" {
		debugApiService.SetP2P(p2ps)
		debugApiService.SetClusterAddress(&clusterAddress)
		debugApiService.MountDebug(false)

		debugApiService.Configure(signer, authenticator, tracer, api.Options{
			CORSAllowedOrigins: o.CORSAllowedOrigins,
			WsPingPeriod:       60 * time.Second,
			Restricted:         o.Restricted,
		}, debugOpts, 1, erc20)
	}

	apiListener, err := net.Listen("tcp", o.APIAddr)
	if err != nil {
		return nil, fmt.Errorf("api listener: %w", err)
	}

	apiServer := &http.Server{
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           apiService,
		ErrorLog:          stdlog.New(b.errorLogWriter, "", 0),
	}

	go func() {
		logger.Info("starting api server", "address", apiListener.Addr())

		if err := apiServer.Serve(apiListener); err != nil && err != http.ErrServerClosed {
			logger.Debug("api server failed to start", "error", err)
			logger.Error(nil, "api server failed to start")
		}
	}()

	b.apiServer = apiServer
	b.apiCloser = apiService

	return b, nil
}

func (b *DevMop) Shutdown() error {
	var mErr error

	tryClose := func(c io.Closer, errMsg string) {
		if c == nil {
			return
		}
		if err := c.Close(); err != nil {
			mErr = multierror.Append(mErr, fmt.Errorf("%s: %w", errMsg, err))
		}
	}

	tryClose(b.apiCloser, "api")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var eg errgroup.Group
	if b.apiServer != nil {
		eg.Go(func() error {
			if err := b.apiServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("api server: %w", err)
			}
			return nil
		})
	}
	if b.debugAPIServer != nil {
		eg.Go(func() error {
			if err := b.debugAPIServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("debug api server: %w", err)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		mErr = multierror.Append(mErr, err)
	}

	tryClose(b.pssCloser, "psser")
	tryClose(b.tracerCloser, "tracer")
	tryClose(b.tagsCloser, "tag persistence")
	tryClose(b.stateStoreCloser, "statestore")
	tryClose(b.localstoreCloser, "localstore")

	return mErr
}

func pong(ctx context.Context, address cluster.Address, msgs ...string) (rtt time.Duration, err error) {
	return time.Millisecond, nil
}
