// Package node defines the concept of a Mop node
// by bootstrapping and injecting all necessary
// dependencies.
package node

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redesblock/mop/core/incentives/pledge"
	"github.com/redesblock/mop/core/incentives/reward"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-multierror"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/auth"
	"github.com/redesblock/mop/core/chain/config"
	chainsyncer "github.com/redesblock/mop/core/chain/syncer"
	"github.com/redesblock/mop/core/chain/transaction"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds/factory"
	"github.com/redesblock/mop/core/incentives/bookkeeper"
	"github.com/redesblock/mop/core/incentives/settlement/swap"
	"github.com/redesblock/mop/core/incentives/settlement/swap/chequebook"
	"github.com/redesblock/mop/core/incentives/settlement/swap/erc20"
	"github.com/redesblock/mop/core/incentives/settlement/swap/priceoracle"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/incentives/voucher/batchservice"
	"github.com/redesblock/mop/core/incentives/voucher/batchstore"
	"github.com/redesblock/mop/core/incentives/voucher/listener"
	"github.com/redesblock/mop/core/incentives/voucher/vouchercontract"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/metrics"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/p2p/libp2p"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/p2p/topology/depthmonitor"
	"github.com/redesblock/mop/core/p2p/topology/kademlia"
	"github.com/redesblock/mop/core/p2p/topology/lightnode"
	"github.com/redesblock/mop/core/pins"
	"github.com/redesblock/mop/core/pricer"
	"github.com/redesblock/mop/core/protocol/chainsync"
	"github.com/redesblock/mop/core/protocol/hive"
	"github.com/redesblock/mop/core/protocol/pingpong"
	"github.com/redesblock/mop/core/protocol/pricing"
	"github.com/redesblock/mop/core/protocol/pseudosettle"
	"github.com/redesblock/mop/core/protocol/pullsync"
	"github.com/redesblock/mop/core/protocol/pullsync/pullstorage"
	"github.com/redesblock/mop/core/protocol/pushsync"
	"github.com/redesblock/mop/core/protocol/retrieval"
	"github.com/redesblock/mop/core/psser"
	"github.com/redesblock/mop/core/puller"
	"github.com/redesblock/mop/core/pusher"
	"github.com/redesblock/mop/core/resolver/multiresolver"
	"github.com/redesblock/mop/core/storer/localstore"
	"github.com/redesblock/mop/core/storer/netstore"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/tracer"
	"github.com/redesblock/mop/core/traverser"
	"github.com/redesblock/mop/core/util"
	"github.com/redesblock/mop/core/util/ioutil"
	"github.com/redesblock/mop/core/warden"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/errgroup"
)

// LoggerName is the tree path name of the logger for this package.
const LoggerName = "node"

type Mop struct {
	p2pService               io.Closer
	p2pHalter                p2p.Halter
	p2pCancel                context.CancelFunc
	apiCloser                []io.Closer
	apiServer                []*http.Server
	debugAPIServer           *http.Server
	resolverCloser           io.Closer
	errorLogWriter           io.Writer
	tracerCloser             io.Closer
	tagsCloser               io.Closer
	stateStoreCloser         io.Closer
	localstoreCloser         io.Closer
	nsCloser                 io.Closer
	topologyCloser           io.Closer
	topologyHalter           topology.Halter
	pusherCloser             io.Closer
	pullerCloser             io.Closer
	accountingCloser         io.Closer
	pullSyncCloser           io.Closer
	pssCloser                io.Closer
	ethClientCloser          func()
	transactionMonitorCloser io.Closer
	transactionCloser        io.Closer
	listenerCloser           io.Closer
	voucherServiceCloser     io.Closer
	priceOracleCloser        io.Closer
	hiveCloser               io.Closer
	chainSyncerCloser        io.Closer
	depthMonitorCloser       io.Closer
	shutdownInProgress       bool
	shutdownMutex            sync.Mutex
	syncingStopped           *util.Signaler
}

type Options struct {
	DataDir                    string
	CacheCapacity              uint64
	MemCacheCapacity           uint64
	DBOpenFilesLimit           uint64
	DBWriteBufferSize          uint64
	DBBlockCacheCapacity       uint64
	DBDisableSeeksCompaction   bool
	APIAddr                    []string
	DebugAPIAddr               string
	Addr                       string
	NATAddr                    string
	EnableWS                   bool
	WelcomeMessage             string
	Bootnodes                  []string
	CORSAllowedOrigins         []string
	Logger                     log.Logger
	TracingEnabled             bool
	TracingEndpoint            string
	TracingServiceName         string
	PaymentThreshold           string
	PaymentTolerance           int64
	PaymentEarly               int64
	ResolverConnectionCfgs     []multiresolver.ConnectionConfig
	RetrievalCaching           bool
	BootnodeMode               bool
	BSCEndpoints               []string
	SwapFactoryAddress         string
	SwapLegacyFactoryAddresses []string
	SwapInitialDeposit         string
	SwapEnable                 bool
	ChequebookEnable           bool
	FullNodeMode               bool
	Transaction                string
	BlockHash                  string
	VoucherContractAddress     string
	PriceOracleAddress         string
	PledgeAddress              string
	RewardAddress              string
	BlockTime                  uint64
	DeployGasPrice             string
	WarmupTime                 time.Duration
	ChainID                    int64
	Resync                     bool
	BlockProfile               bool
	MutexProfile               bool
	StaticNodes                []cluster.Address
	AllowPrivateCIDRs          bool
	Restricted                 bool
	TokenEncryptionKey         string
	AdminPasswordHash          string
	UseVoucherSnapshot         bool
	RemoteEndPoint             string
	TrustNode                  bool
	TLSCertFile                string
	TLSKeyFile                 string
}

func (cfg Options) KeyFile() string {
	path := cfg.TLSKeyFile
	if filepath.IsAbs(path) {
		return path
	}
	return rootify(path, cfg.DataDir)
}

func (cfg Options) CertFile() string {
	path := cfg.TLSCertFile
	if filepath.IsAbs(path) {
		return path
	}
	return rootify(path, cfg.DataDir)
}

func (cfg Options) IsTLSEnabled() bool {
	return cfg.TLSCertFile != "" && cfg.TLSKeyFile != ""
}

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

const (
	refreshRate      = int64(4500000)            // bookkeeper units refreshed per second
	lightFactor      = 10                        // downscale payment thresholds and their change rate, and refresh rates by this for light nodes
	lightRefreshRate = refreshRate / lightFactor // refresh rate used by / for light nodes
	// basePrice                     = 10000                     // minimal price for retrieval and pushsync requests of maximum proximity
	basePrice                     = 0                // TODO REMOVED
	voucherSyncingStallingTimeout = 10 * time.Minute //
	voucherSyncingBackoffTimeout  = 5 * time.Second  //
	minPaymentThreshold           = 2 * refreshRate  // minimal accepted payment threshold of full nodes
	maxPaymentThreshold           = 24 * refreshRate // maximal accepted payment threshold of full nodes
	mainnetNetworkID              = uint64(1)        //
)

func NewMop(interrupt chan struct{}, sysInterrupt chan os.Signal, addr string, publicKey *ecdsa.PublicKey, signer crypto.Signer, networkID uint64, logger log.Logger, libp2pPrivateKey, pssPrivateKey *ecdsa.PrivateKey, o *Options) (b *Mop, err error) {
	tracer, tracerCloser, err := tracer.NewTracer(&tracer.Options{
		Enabled:     o.TracingEnabled,
		Endpoint:    o.TracingEndpoint,
		ServiceName: o.TracingServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("tracer: %w", err)
	}

	p2pCtx, p2pCancel := context.WithCancel(context.Background())
	defer func() {
		// if there's been an error on this function
		// we'd like to cancel the p2p context so that
		// incoming connections will not be possible
		if err != nil {
			p2pCancel()
		}
	}()

	// light nodes have zero warmup time for pull/pushsync protocols
	warmupTime := o.WarmupTime
	if !o.FullNodeMode {
		warmupTime = 0
	}

	sink := ioutil.WriterFunc(func(p []byte) (int, error) {
		logger.Error(nil, string(p))
		return len(p), nil
	})

	b = &Mop{
		p2pCancel:      p2pCancel,
		errorLogWriter: sink,
		tracerCloser:   tracerCloser,
		syncingStopped: util.NewSignaler(),
	}

	defer func(b *Mop) {
		if err != nil {
			logger.Error(err, "got error, shutting down...")
			if err2 := b.Shutdown(); err2 != nil {
				logger.Error(err2, "got error while shutting down")
			}
		}
	}(b)

	stateStore, err := InitStateStore(logger, o.DataDir)
	if err != nil {
		return nil, err
	}
	b.stateStoreCloser = stateStore

	// Check if the the batchstore exists. If not, we can assume it's missing
	// due to a migration or it's a fresh install.
	batchStoreExists, err := batchStoreExists(stateStore)
	if err != nil {
		return nil, fmt.Errorf("batchstore: exists: %w", err)
	}

	addressbook := address.New(stateStore)

	var (
		chainBackend       transaction.Backend
		overlayEthAddress  common.Address
		chainID            int64
		transactionService transaction.Service
		transactionMonitor transaction.Monitor
		chequebookFactory  chequebook.Factory
		chequebookService  chequebook.Service = new(noOpChequebookService)
		chequeStore        chequebook.ChequeStore
		cashoutService     chequebook.CashoutService
		pollingInterval    = time.Duration(o.BlockTime) * time.Second
		erc20Service       erc20.Service
	)

	chainEnabled := isChainEnabled(o, o.BSCEndpoints, logger)

	var batchStore voucher.Storer = new(voucher.NoOpBatchStore)
	var unreserveFn func([]byte, uint8) (uint64, error)

	if chainEnabled {
		var evictFn = func(b []byte) error {
			_, err := unreserveFn(b, cluster.MaxPO+1)
			return err
		}
		batchStore, err = batchstore.New(stateStore, evictFn, logger)
		if err != nil {
			return nil, fmt.Errorf("batchstore: %w", err)
		}
	}

	chainBackend, overlayEthAddress, chainID, transactionMonitor, transactionService, err = InitChain(
		p2pCtx,
		logger,
		stateStore,
		o.BSCEndpoints,
		o.ChainID,
		signer,
		pollingInterval,
		chainEnabled)
	if err != nil {
		return nil, fmt.Errorf("init chain: %w", err)
	}
	b.ethClientCloser = chainBackend.Close

	if o.ChainID != -1 && o.ChainID != chainID {
		return nil, fmt.Errorf("connected to wrong BNB Smart Chain network; network chainID %d; configured chainID %d", chainID, o.ChainID)
	}

	b.transactionCloser = tracerCloser
	b.transactionMonitorCloser = transactionMonitor

	var authenticator *auth.Authenticator

	if o.Restricted {
		if authenticator, err = auth.New(o.TokenEncryptionKey, o.AdminPasswordHash, logger); err != nil {
			return nil, fmt.Errorf("authenticator: %w", err)
		}
		logger.Info("starting with restricted APIs")
	}

	// set up basic debug api endpoints for debugging and /health endpoint
	mopNodeMode := api.LightMode
	if o.FullNodeMode {
		mopNodeMode = api.FullMode
	} else if !chainEnabled {
		mopNodeMode = api.UltraLightMode
	}

	probe := api.NewProbe()
	probe.SetHealthy(api.ProbeStatusOK)
	defer func(probe *api.Probe) {
		if err != nil {
			probe.SetHealthy(api.ProbeStatusNOK)
		} else {
			probe.SetReady(api.ProbeStatusOK)
		}
	}(probe)

	var debugService *api.Service

	if o.DebugAPIAddr != "" {
		if o.MutexProfile {
			_ = runtime.SetMutexProfileFraction(1)
		}

		if o.BlockProfile {
			runtime.SetBlockProfileRate(1)
		}

		debugAPIListener, err := net.Listen("tcp", o.DebugAPIAddr)
		if err != nil {
			return nil, fmt.Errorf("debug api listener: %w", err)
		}

		debugService = api.New(*publicKey, pssPrivateKey.PublicKey, overlayEthAddress, logger, transactionService, batchStore, mopNodeMode, o.ChequebookEnable, o.SwapEnable, chainBackend, o.CORSAllowedOrigins)
		debugService.MountTechnicalDebug()
		debugService.SetProbe(probe)

		debugAPIServer := &http.Server{
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           debugService,
			ErrorLog:          stdlog.New(b.errorLogWriter, "", 0),
		}

		go func() {
			logger.Info("starting debug server", "address", debugAPIListener.Addr())

			if o.IsTLSEnabled() {
				if err := debugAPIServer.ServeTLS(debugAPIListener, o.CertFile(), o.KeyFile()); err != nil && err != http.ErrServerClosed {
					logger.Debug("debug api server failed to start", "error", err)
					logger.Error(nil, "debug api server failed to start")
				}
			} else {
				if err := debugAPIServer.Serve(debugAPIListener); err != nil && err != http.ErrServerClosed {
					logger.Debug("debug api server failed to start", "error", err)
					logger.Error(nil, "debug api server failed to start")
				}
			}
		}()

		b.debugAPIServer = debugAPIServer
	}

	var apiService *api.Service

	if o.Restricted {
		apiService = api.New(*publicKey, pssPrivateKey.PublicKey, overlayEthAddress, logger, transactionService, batchStore, mopNodeMode, o.ChequebookEnable, o.SwapEnable, chainBackend, o.CORSAllowedOrigins)
		apiService.MountTechnicalDebug()

		apiServer := &http.Server{
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           apiService,
			ErrorLog:          stdlog.New(b.errorLogWriter, "", 0),
		}

		for _, apiAddr := range o.APIAddr {
			apiListener, err := net.Listen("tcp", apiAddr)
			if err != nil {
				return nil, fmt.Errorf("api listener: %w", err)
			}

			go func() {
				logger.Info("starting debug & api server", "address", apiListener.Addr())

				if o.IsTLSEnabled() {
					if err := apiServer.ServeTLS(apiListener, o.CertFile(), o.KeyFile()); err != nil && err != http.ErrServerClosed {
						logger.Debug("debug & api server failed to start", "error", err)
						logger.Error(nil, "debug & api server failed to start")
					}
				} else {
					if err := apiServer.Serve(apiListener); err != nil && err != http.ErrServerClosed {
						logger.Debug("debug & api server failed to start", "error", err)
						logger.Error(nil, "debug & api server failed to start")
					}
				}
			}()

			b.apiServer = append(b.apiServer, apiServer)
			b.apiCloser = append(b.apiCloser, apiServer)

		}
	}

	// Sync the with the given BNB Smart Chain backend:
	isSynced, _, err := transaction.IsSynced(p2pCtx, chainBackend, maxDelay)
	if err != nil {
		return nil, fmt.Errorf("is synced: %w", err)
	}
	if !isSynced {
		logger.Info("waiting to chainsync with the BNB Smart Chain backend")

		err := transaction.WaitSynced(p2pCtx, logger, chainBackend, maxDelay)
		if err != nil {
			return nil, fmt.Errorf("waiting backend chainsync: %w", err)
		}
	}

	if o.SwapEnable {
		chequebookFactory, err = InitChequebookFactory(
			logger,
			chainBackend,
			chainID,
			transactionService,
			o.SwapFactoryAddress,
			o.SwapLegacyFactoryAddresses,
		)
		if err != nil {
			return nil, err
		}

		if err = chequebookFactory.VerifyBytecode(p2pCtx); err != nil {
			return nil, fmt.Errorf("factory fail: %w", err)
		}

		erc20Address, err := chequebookFactory.ERC20Address(p2pCtx)
		if err != nil {
			return nil, fmt.Errorf("factory fail: %w", err)
		}

		erc20Service = erc20.New(transactionService, erc20Address)

		if o.ChequebookEnable && chainEnabled {
			chequebookService, err = InitChequebookService(
				p2pCtx,
				logger,
				stateStore,
				signer,
				chainID,
				chainBackend,
				overlayEthAddress,
				transactionService,
				chequebookFactory,
				o.SwapInitialDeposit,
				o.DeployGasPrice,
				erc20Service,
			)
			if err != nil {
				return nil, err
			}
		}

		chequeStore, cashoutService = initChequeStoreCashout(
			stateStore,
			chainBackend,
			chequebookFactory,
			chainID,
			overlayEthAddress,
			transactionService,
		)
	}

	pubKey, _ := signer.PublicKey()
	if err != nil {
		return nil, err
	}

	// if there's a previous transaction hash, and not a new chequebook deployment on a node starting from scratch
	// get old overlay
	// mine nonce that gives similar new overlay
	nonce, nonceExists, err := overlayNonceExists(stateStore)
	if err != nil {
		return nil, fmt.Errorf("check presence of nonce: %w", err)
	}
	if !nonceExists {
		nonce = make([]byte, 32)
	}

	existingOverlay, err := GetExistingOverlay(stateStore)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("get existing overlay: %w", err)
		}
		logger.Warning("existing overlay", "error", err)
	}

	if err == nil && o.FullNodeMode && !nonceExists {
		newOverlayCandidate := cluster.ZeroAddress
		j := uint64(0)
		limit := math.Pow(2, 34)
		for prox := uint8(0); prox < cluster.MaxPO && j < uint64(limit); j++ {
			select {
			case <-sysInterrupt:
				return nil, errors.New("interrupted while finding new overlay")
			default:
			}
			binary.LittleEndian.PutUint64(nonce, j)
			if (j/1000000)*1000000 == j {
				logger.Debug("finding new overlay corresponding to previous overlay with nonce", "nonce", hex.EncodeToString(nonce))
			}
			newOverlayCandidate, err = crypto.NewOverlayAddress(*pubKey, networkID, nonce)
			if err == nil {
				prox = cluster.Proximity(existingOverlay.Bytes(), newOverlayCandidate.Bytes())
			} else {
				logger.Debug("error finding new overlay", "nonce", hex.EncodeToString(nonce), "error", err)
			}
		}

		foundProximity := cluster.Proximity(existingOverlay.Bytes(), newOverlayCandidate.Bytes())
		if foundProximity < cluster.MaxPO {
			return nil, fmt.Errorf("mining new overlay address failed")
		}
	}

	clusterAddress, err := crypto.NewOverlayAddress(*pubKey, networkID, nonce)
	if err != nil {
		return nil, fmt.Errorf("compute overlay address: %w", err)
	}
	logger.Info("using overlay address", "address", clusterAddress)

	if !nonceExists {
		err := setOverlayNonce(stateStore, nonce)
		if err != nil {
			return nil, fmt.Errorf("statestore: save new overlay nonce: %w", err)
		}

		err = SetOverlayInStore(clusterAddress, stateStore)
		if err != nil {
			return nil, fmt.Errorf("statestore: save new overlay: %w", err)
		}
	}

	apiService.SetClusterAddress(&clusterAddress)

	if err = CheckOverlayWithStore(clusterAddress, stateStore); err != nil {
		return nil, err
	}

	lightNodes := lightnode.NewContainer(clusterAddress)

	var bootnodes []ma.Multiaddr

	for _, a := range o.Bootnodes {
		addr, err := ma.NewMultiaddr(a)
		if err != nil {
			logger.Debug("create bootnode multiaddress from string failed", "string", a, "error", err)
			logger.Warning("create bootnode multiaddress from string failed", "string", a)
			continue
		}

		bootnodes = append(bootnodes, addr)
	}

	// Perform checks related to payment threshold calculations here to not duplicate
	// the checks in bootstrap process
	paymentThreshold, ok := new(big.Int).SetString(o.PaymentThreshold, 10)
	if !ok {
		return nil, fmt.Errorf("invalid payment threshold: %s", paymentThreshold)
	}

	if paymentThreshold.Cmp(big.NewInt(minPaymentThreshold)) < 0 {
		return nil, fmt.Errorf("payment threshold below minimum generally accepted value, need at least %d", minPaymentThreshold)
	}

	if paymentThreshold.Cmp(big.NewInt(maxPaymentThreshold)) > 0 {
		return nil, fmt.Errorf("payment threshold above maximum generally accepted value, needs to be reduced to at most %d", maxPaymentThreshold)
	}

	if o.PaymentTolerance < 0 {
		return nil, fmt.Errorf("invalid payment tolerance: %d", o.PaymentTolerance)
	}

	if o.PaymentEarly > 100 || o.PaymentEarly < 0 {
		return nil, fmt.Errorf("invalid payment early: %d", o.PaymentEarly)
	}

	var initBatchState *voucher.ChainSnapshot
	// Bootstrap node with voucher snapshot only if it is running on mainnet, is a fresh
	// install or explicitly asked by user to resync
	if networkID == mainnetNetworkID && o.UseVoucherSnapshot && (!batchStoreExists || o.Resync) {
		start := time.Now()
		logger.Info("cold voucher start detected. fetching voucher stamp snapshot from cluster")
		initBatchState, err = bootstrapNode(
			addr,
			clusterAddress,
			nonce,
			chainID,
			overlayEthAddress,
			addressbook,
			bootnodes,
			lightNodes,
			chequebookService,
			chequeStore,
			cashoutService,
			transactionService,
			stateStore,
			signer,
			networkID,
			log.Noop,
			libp2pPrivateKey,
			o,
		)
		logger.Info("bootstrapper created", "elapsed", time.Since(start))
		if err != nil {
			logger.Error(err, "bootstrapper failed to fetch batch state")
		}
	}

	p2ps, err := libp2p.New(p2pCtx, signer, networkID, clusterAddress, addr, addressbook, stateStore, lightNodes, logger, tracer, libp2p.Options{
		PrivateKey:      libp2pPrivateKey,
		NATAddr:         o.NATAddr,
		EnableWS:        o.EnableWS,
		WelcomeMessage:  o.WelcomeMessage,
		FullNode:        o.FullNodeMode,
		Nonce:           nonce,
		ValidateOverlay: chainEnabled,
	})
	if err != nil {
		return nil, fmt.Errorf("p2p service: %w", err)
	}

	apiService.SetP2P(p2ps)

	b.p2pService = p2ps
	b.p2pHalter = p2ps

	// localstore depends on batchstore
	var path string

	if o.DataDir != "" {
		logger.Info("using datadir", "path", o.DataDir)
		path = filepath.Join(o.DataDir, "localstore")
	}
	lo := &localstore.Options{
		Capacity:               o.CacheCapacity,
		MemCapacity:            o.MemCacheCapacity,
		ReserveCapacity:        uint64(batchstore.Capacity),
		UnreserveFunc:          batchStore.Unreserve,
		OpenFilesLimit:         o.DBOpenFilesLimit,
		BlockCacheCapacity:     o.DBBlockCacheCapacity,
		WriteBufferSize:        o.DBWriteBufferSize,
		DisableSeeksCompaction: o.DBDisableSeeksCompaction,
	}

	storer, err := localstore.New(path, clusterAddress.Bytes(), stateStore, lo, logger)
	if err != nil {
		return nil, fmt.Errorf("localstore: %w", err)
	}
	b.localstoreCloser = storer
	unreserveFn = storer.UnreserveBatch

	validStamp := voucher.ValidStamp(batchStore)
	post, err := voucher.NewService(stateStore, batchStore, chainID)
	if err != nil {
		return nil, fmt.Errorf("voucher service load: %w", err)
	}
	b.voucherServiceCloser = post
	batchStore.SetBatchExpiryHandler(post)

	var (
		voucherContractService vouchercontract.Interface
		batchSvc               voucher.EventUpdater
		eventListener          voucher.Listener
	)

	var voucherSyncStart uint64 = 0

	chainCfg, found := config.GetChainConfig(chainID)
	voucherContractAddress, startBlock := chainCfg.VoucherStamp, chainCfg.StartBlock
	if o.VoucherContractAddress != "" {
		if !common.IsHexAddress(o.VoucherContractAddress) {
			return nil, errors.New("malformed voucher stamp address")
		}
		voucherContractAddress = common.HexToAddress(o.VoucherContractAddress)
	} else if !found {
		return nil, errors.New("no known voucher stamp addresses for this network")
	}
	if found {
		voucherSyncStart = startBlock
	}

	eventListener = listener.New(b.syncingStopped, logger, chainBackend, voucherContractAddress, o.BlockTime, voucherSyncingStallingTimeout, voucherSyncingBackoffTimeout)
	b.listenerCloser = eventListener

	batchSvc, err = batchservice.New(stateStore, batchStore, logger, eventListener, overlayEthAddress.Bytes(), post, sha3.New256, o.Resync)
	if err != nil {
		return nil, fmt.Errorf("batch service load: %w", err)
	}

	erc20Address, err := vouchercontract.LookupERC20Address(p2pCtx, transactionService, voucherContractAddress, chainEnabled)
	if err != nil {
		return nil, fmt.Errorf("vouchercontract lookup erc20: %w", err)
	}

	voucherContractService = vouchercontract.New(
		overlayEthAddress,
		voucherContractAddress,
		erc20Address,
		transactionService,
		post,
		batchStore,
		chainEnabled,
	)

	pledgeAddress := chainCfg.PledgeAddress
	if o.PledgeAddress != "" {
		if !common.IsHexAddress(o.PledgeAddress) {
			return nil, errors.New("malformed pledge address")
		}
		pledgeAddress = common.HexToAddress(o.PledgeAddress)
	} else if !found {
		return nil, errors.New("no known pledge addresses for this network")
	}

	pledgeContractService := pledge.New(
		stateStore,
		transactionService,
		pledgeAddress,
	)

	rewardAddress := chainCfg.RewardAddress
	if o.RewardAddress != "" {
		if !common.IsHexAddress(o.RewardAddress) {
			return nil, errors.New("malformed reward address")
		}
		rewardAddress = common.HexToAddress(o.RewardAddress)
	} else if !found {
		return nil, errors.New("no known reward addresses for this network")
	}
	rewardContractService := reward.New(
		stateStore,
		transactionService,
		rewardAddress,
	)

	if natManager := p2ps.NATManager(); natManager != nil {
		// wait for nat manager to init
		logger.Debug("initializing NAT manager")
		select {
		case <-natManager.Ready():
			// this is magic sleep to give NAT time to chainsync the mappings
			// this is a hack, kind of alchemy and should be improved
			time.Sleep(3 * time.Second)
			logger.Debug("NAT manager initialized")
		case <-time.After(10 * time.Second):
			logger.Warning("NAT manager init timeout")
		}
	}

	// Construct protocols.
	pingPong := pingpong.New(p2ps, logger, tracer)

	if err = p2ps.AddProtocol(pingPong.Protocol()); err != nil {
		return nil, fmt.Errorf("pingpong service: %w", err)
	}

	hive, err := hive.New(p2ps, addressbook, networkID, o.BootnodeMode, o.AllowPrivateCIDRs, logger)
	if err != nil {
		return nil, fmt.Errorf("hive: %w", err)
	}

	if err = p2ps.AddProtocol(hive.Protocol()); err != nil {
		return nil, fmt.Errorf("hive service: %w", err)
	}
	b.hiveCloser = hive

	var swapService *swap.Service

	metricsDB, err := shed.NewDBWrap(stateStore.DB())
	if err != nil {
		return nil, fmt.Errorf("unable to create metrics storage for kademlia: %w", err)
	}

	kad, err := kademlia.New(clusterAddress, addressbook, hive, p2ps, pingPong, metricsDB, logger,
		kademlia.Options{Bootnodes: bootnodes, BootnodeMode: o.BootnodeMode, StaticNodes: o.StaticNodes, IgnoreRadius: !chainEnabled})
	if err != nil {
		return nil, fmt.Errorf("unable to create kademlia: %w", err)
	}
	b.topologyCloser = kad
	b.topologyHalter = kad
	hive.SetAddPeersHandler(kad.AddPeers)
	p2ps.SetPickyNotifier(kad)

	var (
		syncErr    atomic.Value
		syncStatus atomic.Value

		syncStatusFn = func() (isDone bool, err error) {
			iErr := syncErr.Load()
			if iErr != nil {
				err = iErr.(error)
			}
			isDone = syncStatus.Load() != nil
			return isDone, err
		}
	)

	if batchSvc != nil && chainEnabled {
		logger.Info("waiting to chainsync voucher contract data, this may take a while... more info available in Debug loglevel")
		if o.FullNodeMode {
			err = batchSvc.Start(voucherSyncStart, initBatchState, interrupt)
			syncStatus.Store(true)
			if err != nil {
				syncErr.Store(err)
				return nil, fmt.Errorf("unable to start batch service: %w", err)
			} else {
				err = post.SetExpired()
				if err != nil {
					return nil, fmt.Errorf("unable to set expirations: %w", err)
				}
			}
		} else {
			go func() {
				logger.Info("started voucher contract data chainsync in the background...")
				err := batchSvc.Start(voucherSyncStart, initBatchState, interrupt)
				syncStatus.Store(true)
				if err != nil {
					syncErr.Store(err)
					logger.Error(err, "unable to chainsync batches")
					b.syncingStopped.Signal() // trigger shutdown in start.go
				} else {
					err = post.SetExpired()
					if err != nil {
						logger.Error(err, "unable to set expirations")
					}
				}
			}()
		}
	}

	minThreshold := big.NewInt(2 * refreshRate)
	maxThreshold := big.NewInt(24 * refreshRate)

	if !o.FullNodeMode {
		minThreshold = big.NewInt(2 * lightRefreshRate)
	}

	lightPaymentThreshold := new(big.Int).Div(paymentThreshold, big.NewInt(lightFactor))

	pricer := pricer.NewFixedPricer(clusterAddress, basePrice)

	if paymentThreshold.Cmp(minThreshold) < 0 {
		return nil, fmt.Errorf("payment threshold below minimum generally accepted value, need at least %s", minThreshold)
	}

	if paymentThreshold.Cmp(maxThreshold) > 0 {
		return nil, fmt.Errorf("payment threshold above maximum generally accepted value, needs to be reduced to at most %s", maxThreshold)
	}

	pricing := pricing.New(p2ps, logger, paymentThreshold, lightPaymentThreshold, minThreshold)

	if err = p2ps.AddProtocol(pricing.Protocol()); err != nil {
		return nil, fmt.Errorf("pricing service: %w", err)
	}

	addrs, err := p2ps.Addresses()
	if err != nil {
		return nil, fmt.Errorf("get server addresses: %w", err)
	}

	for _, addr := range addrs {
		logger.Debug("p2p address", "address", addr)
	}

	var enforcedRefreshRate *big.Int

	if o.FullNodeMode {
		enforcedRefreshRate = big.NewInt(refreshRate)
	} else {
		enforcedRefreshRate = big.NewInt(lightRefreshRate)
	}

	acc, err := bookkeeper.NewAccounting(
		paymentThreshold,
		o.PaymentTolerance,
		o.PaymentEarly,
		logger,
		stateStore,
		pricing,
		new(big.Int).Set(enforcedRefreshRate),
		lightFactor,
		p2ps,
	)
	if err != nil {
		return nil, fmt.Errorf("bookkeeper: %w", err)
	}
	b.accountingCloser = acc

	pseudosettleService := pseudosettle.New(p2ps, logger, stateStore, acc, new(big.Int).Set(enforcedRefreshRate), big.NewInt(lightRefreshRate), p2ps)
	if err = p2ps.AddProtocol(pseudosettleService.Protocol()); err != nil {
		return nil, fmt.Errorf("pseudosettle service: %w", err)
	}

	acc.SetRefreshFunc(pseudosettleService.Pay)

	if o.SwapEnable && chainEnabled {
		var priceOracle priceoracle.Service
		swapService, priceOracle, err = InitSwap(
			p2ps,
			logger,
			stateStore,
			networkID,
			overlayEthAddress,
			chequebookService,
			chequeStore,
			cashoutService,
			acc,
			o.PriceOracleAddress,
			chainID,
			transactionService,
		)
		if err != nil {
			return nil, err
		}
		b.priceOracleCloser = priceOracle

		if o.ChequebookEnable {
			acc.SetPayFunc(swapService.Pay)
		}
	}

	pricing.SetPaymentThresholdObserver(acc)

	retrieve := retrieval.New(clusterAddress, storer, p2ps, kad, logger, acc, pricer, tracer, o.RetrievalCaching, validStamp)
	tagService := tags.NewTags(stateStore, logger)
	b.tagsCloser = tagService

	pssService := psser.New(pssPrivateKey, logger)
	b.pssCloser = pssService

	var netStorer = netstore.New(storer, validStamp, retrieve, logger, o.TrustNode)
	b.nsCloser = netStorer

	traversalService := traverser.New(netStorer)

	pinningService := pins.NewService(storer, stateStore, traversalService)

	pushSyncProtocol := pushsync.New(clusterAddress, nonce, p2ps, storer, kad, tagService, o.FullNodeMode, pssService.TryUnwrap, validStamp, logger, acc, pricer, signer, tracer, warmupTime, o.RemoteEndPoint)

	// set the pushSyncer in the PSS
	pssService.SetPushSyncer(pushSyncProtocol)

	pusherService := pusher.New(networkID, storer, kad, pushSyncProtocol, validStamp, tagService, logger, tracer, warmupTime)
	b.pusherCloser = pusherService

	pullStorage := pullstorage.New(storer)

	pullSyncProtocol := pullsync.New(p2ps, pullStorage, pssService.TryUnwrap, validStamp, logger)
	b.pullSyncCloser = pullSyncProtocol

	var pullerService *puller.Puller
	if o.FullNodeMode && !o.BootnodeMode {
		pullerService = puller.New(stateStore, kad, batchStore, pullSyncProtocol, logger, puller.Options{}, warmupTime)
		b.pullerCloser = pullerService
	}

	retrieveProtocolSpec := retrieve.Protocol()
	pushSyncProtocolSpec := pushSyncProtocol.Protocol()
	pullSyncProtocolSpec := pullSyncProtocol.Protocol()

	if o.FullNodeMode {
		logger.Info("starting in full mode")
	} else {
		logger.Info("starting in light mode")
		p2p.WithBlocklistStreams(p2p.DefaultBlocklistTime, retrieveProtocolSpec)
		p2p.WithBlocklistStreams(p2p.DefaultBlocklistTime, pushSyncProtocolSpec)
		p2p.WithBlocklistStreams(p2p.DefaultBlocklistTime, pullSyncProtocolSpec)
	}

	if err = p2ps.AddProtocol(retrieveProtocolSpec); err != nil {
		return nil, fmt.Errorf("retrieval service: %w", err)
	}
	if err = p2ps.AddProtocol(pushSyncProtocolSpec); err != nil {
		return nil, fmt.Errorf("pushsync service: %w", err)
	}
	if err = p2ps.AddProtocol(pullSyncProtocolSpec); err != nil {
		return nil, fmt.Errorf("pullsync protocol: %w", err)
	}

	if o.FullNodeMode {
		depthMonitor := depthmonitor.New(kad, pullSyncProtocol, storer, batchStore, logger, warmupTime, depthmonitor.DefaultWakeupInterval)
		b.depthMonitorCloser = depthMonitor
	}

	multiResolver := multiresolver.NewMultiResolver(
		multiresolver.WithConnectionConfigs(o.ResolverConnectionCfgs),
		multiresolver.WithLogger(o.Logger),
		multiresolver.WithDefaultCIDResolver(),
	)
	b.resolverCloser = multiResolver
	var chainSyncer *chainsyncer.ChainSyncer

	if o.FullNodeMode {
		cs, err := chainsync.New(p2ps, chainBackend)
		if err != nil {
			return nil, fmt.Errorf("new chainsync: %w", err)
		}
		if err = p2ps.AddProtocol(cs.Protocol()); err != nil {
			return nil, fmt.Errorf("chainsync protocol: %w", err)
		}
		chainSyncer, err = chainsyncer.New(chainBackend, cs, kad, p2ps, logger, nil)
		if err != nil {
			return nil, fmt.Errorf("new chainsyncer: %w", err)
		}

		b.chainSyncerCloser = chainSyncer
	}

	feedFactory := factory.New(netStorer)
	warden := warden.New(storer, traversalService, retrieve, pushSyncProtocol)

	extraOpts := api.ExtraOptions{
		Pingpong:         pingPong,
		TopologyDriver:   kad,
		LightNodes:       lightNodes,
		Accounting:       acc,
		Pseudosettle:     pseudosettleService,
		Swap:             swapService,
		Chequebook:       chequebookService,
		BlockTime:        big.NewInt(int64(o.BlockTime)),
		Tags:             tagService,
		Storer:           netStorer,
		Resolver:         multiResolver,
		Pss:              pssService,
		TraversalService: traversalService,
		Pinning:          pinningService,
		FeedFactory:      feedFactory,
		Post:             post,
		VoucherContract:  voucherContractService,
		PledgeContract:   pledgeContractService,
		RewardContract:   rewardContractService,
		Warden:           warden,
		SyncStatus:       syncStatusFn,
	}

	if len(o.APIAddr) != 0 {
		if apiService == nil {
			apiService = api.New(*publicKey, pssPrivateKey.PublicKey, overlayEthAddress, logger, transactionService, batchStore, mopNodeMode, o.ChequebookEnable, o.SwapEnable, chainBackend, o.CORSAllowedOrigins)
		}

		chunkC := apiService.Configure(signer, authenticator, tracer, api.Options{
			CORSAllowedOrigins: o.CORSAllowedOrigins,
			WsPingPeriod:       60 * time.Second,
			Restricted:         o.Restricted,
			NATAddr:            o.NATAddr,
		}, extraOpts, chainID, erc20Service)

		pusherService.AddFeed(chunkC)

		apiService.MountAPI()
		apiService.SetProbe(probe)

		if !o.Restricted {
			apiServer := &http.Server{
				IdleTimeout:       30 * time.Second,
				ReadHeaderTimeout: 3 * time.Second,
				Handler:           apiService,
				ErrorLog:          stdlog.New(b.errorLogWriter, "", 0),
			}

			for _, apiAddr := range o.APIAddr {
				apiListener, err := net.Listen("tcp", apiAddr)
				if err != nil {
					return nil, fmt.Errorf("api listener: %w", err)
				}

				go func() {
					logger.Info("starting api server", "address", apiListener.Addr())
					if o.IsTLSEnabled() {
						if err := apiServer.ServeTLS(apiListener, o.CertFile(), o.KeyFile()); err != nil && err != http.ErrServerClosed {
							logger.Debug("api server failed to start", "error", err)
							logger.Error(nil, "api server failed to start")
						}
					} else {
						if err := apiServer.Serve(apiListener); err != nil && err != http.ErrServerClosed {
							logger.Debug("api server failed to start", "error", err)
							logger.Error(nil, "api server failed to start")
						}
					}
				}()

				b.apiServer = append(b.apiServer, apiServer)
				b.apiCloser = append(b.apiCloser, apiServer)
			}
		} else {
			// in Restricted mode we mount debug endpoints
			apiService.MountDebug(o.Restricted)
		}
	}

	if o.DebugAPIAddr != "" {
		// register metrics from components
		debugService.MustRegisterMetrics(p2ps.Metrics()...)
		debugService.MustRegisterMetrics(pingPong.Metrics()...)
		debugService.MustRegisterMetrics(acc.Metrics()...)
		debugService.MustRegisterMetrics(storer.Metrics()...)
		debugService.MustRegisterMetrics(kad.Metrics()...)

		if pullerService != nil {
			debugService.MustRegisterMetrics(pullerService.Metrics()...)
		}

		debugService.MustRegisterMetrics(pushSyncProtocol.Metrics()...)
		debugService.MustRegisterMetrics(pusherService.Metrics()...)
		debugService.MustRegisterMetrics(pullSyncProtocol.Metrics()...)
		debugService.MustRegisterMetrics(pullStorage.Metrics()...)
		debugService.MustRegisterMetrics(retrieve.Metrics()...)
		debugService.MustRegisterMetrics(lightNodes.Metrics()...)
		debugService.MustRegisterMetrics(hive.Metrics()...)

		if bs, ok := batchStore.(metrics.Collector); ok {
			debugService.MustRegisterMetrics(bs.Metrics()...)
		}
		if eventListener != nil {
			if ls, ok := eventListener.(metrics.Collector); ok {
				debugService.MustRegisterMetrics(ls.Metrics()...)
			}
		}
		if pssServiceMetrics, ok := pssService.(metrics.Collector); ok {
			debugService.MustRegisterMetrics(pssServiceMetrics.Metrics()...)
		}
		if swapBackendMetrics, ok := chainBackend.(metrics.Collector); ok {
			debugService.MustRegisterMetrics(swapBackendMetrics.Metrics()...)
		}
		if apiService != nil {
			debugService.MustRegisterMetrics(apiService.Metrics()...)
		}
		if l, ok := logger.(metrics.Collector); ok {
			debugService.MustRegisterMetrics(l.Metrics()...)
		}
		if nsMetrics, ok := netStorer.(metrics.Collector); ok {
			debugService.MustRegisterMetrics(nsMetrics.Metrics()...)
		}
		debugService.MustRegisterMetrics(pseudosettleService.Metrics()...)
		if swapService != nil {
			debugService.MustRegisterMetrics(swapService.Metrics()...)
		}
		if chainSyncer != nil {
			debugService.MustRegisterMetrics(chainSyncer.Metrics()...)
		}

		debugService.Configure(signer, authenticator, tracer, api.Options{
			CORSAllowedOrigins: o.CORSAllowedOrigins,
			WsPingPeriod:       60 * time.Second,
			Restricted:         o.Restricted,
			NATAddr:            o.NATAddr,
		}, extraOpts, chainID, erc20Service)

		debugService.SetP2P(p2ps)
		debugService.SetClusterAddress(&clusterAddress)
		debugService.MountDebug(false)
	}

	if err := kad.Start(p2pCtx); err != nil {
		return nil, err
	}

	if err := p2ps.Ready(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Mop) SyncingStopped() chan struct{} {
	return b.syncingStopped.C
}

func (b *Mop) Shutdown() error {
	var mErr error

	// if a shutdown is already in process, return here
	b.shutdownMutex.Lock()
	if b.shutdownInProgress {
		b.shutdownMutex.Unlock()
		return ErrShutdownInProgress
	}
	b.shutdownInProgress = true
	b.shutdownMutex.Unlock()

	// halt kademlia while shutting down other
	// components.
	if b.topologyHalter != nil {
		b.topologyHalter.Halt()
	}

	// halt p2p layer from accepting new connections
	// while shutting down other components
	if b.p2pHalter != nil {
		b.p2pHalter.Halt()
	}
	// tryClose is a convenient closure which decrease
	// repetitive io.Closer tryClose procedure.
	tryClose := func(c io.Closer, errMsg string) {
		if c == nil {
			return
		}
		if err := c.Close(); err != nil {
			mErr = multierror.Append(mErr, fmt.Errorf("%s: %w", errMsg, err))
		}
	}

	for _, apiCloser := range b.apiCloser {
		tryClose(apiCloser, "api")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var eg errgroup.Group
	if len(b.apiServer) != 0 {
		for _, apiServer := range b.apiServer {
			eg.Go(func() error {
				if err := apiServer.Shutdown(ctx); err != nil {
					return fmt.Errorf("api server: %w", err)
				}
				return nil
			})
		}
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

	var wg sync.WaitGroup
	wg.Add(7)
	go func() {
		defer wg.Done()
		tryClose(b.chainSyncerCloser, "chain syncer")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.pssCloser, "psser")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.pusherCloser, "pusher")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.pullerCloser, "puller")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.accountingCloser, "bookkeeper")
	}()

	b.p2pCancel()
	go func() {
		defer wg.Done()
		tryClose(b.pullSyncCloser, "pull chainsync")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.hiveCloser, "hive")
	}()

	wg.Wait()

	tryClose(b.p2pService, "p2p server")
	tryClose(b.priceOracleCloser, "price oracle service")

	wg.Add(3)
	go func() {
		defer wg.Done()
		tryClose(b.transactionMonitorCloser, "transaction monitor")
		tryClose(b.transactionCloser, "transaction")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.listenerCloser, "listener")
	}()
	go func() {
		defer wg.Done()
		tryClose(b.voucherServiceCloser, "voucher service")
	}()

	wg.Wait()

	if c := b.ethClientCloser; c != nil {
		c()
	}

	tryClose(b.tracerCloser, "tracer")
	tryClose(b.tagsCloser, "tag persistence")
	tryClose(b.topologyCloser, "topology driver")
	tryClose(b.nsCloser, "netstore")
	tryClose(b.depthMonitorCloser, "depthmonitor service")
	tryClose(b.stateStoreCloser, "statestore")
	tryClose(b.localstoreCloser, "localstore")
	tryClose(b.resolverCloser, "resolver service")

	return mErr
}

var ErrShutdownInProgress error = errors.New("shutdown in progress")

func isChainEnabled(o *Options, bscEndpoints []string, logger log.Logger) bool {
	chainDisabled := len(bscEndpoints) == 0
	lightMode := !o.FullNodeMode

	if lightMode && chainDisabled { // ultra light mode is LightNode mode with chain disabled
		logger.Info("starting with a disabled chain backend")
		return false
	}

	logger.Info("starting with an enabled chain backend")
	return true // all other modes operate require chain enabled
}
