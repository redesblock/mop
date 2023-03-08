package node

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-multierror"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/redesblock/mop/core/address"
	"github.com/redesblock/mop/core/chain/transaction"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/feeds/factory"
	"github.com/redesblock/mop/core/file"
	"github.com/redesblock/mop/core/file/joiner"
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/incentives/bookkeeper"
	"github.com/redesblock/mop/core/incentives/settlement/swap/chequebook"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/manifest"
	"github.com/redesblock/mop/core/p2p/libp2p"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/p2p/topology/kademlia"
	"github.com/redesblock/mop/core/p2p/topology/lightnode"
	"github.com/redesblock/mop/core/pricer"
	"github.com/redesblock/mop/core/protocol/hive"
	"github.com/redesblock/mop/core/protocol/pricing"
	"github.com/redesblock/mop/core/protocol/pseudosettle"
	"github.com/redesblock/mop/core/protocol/retrieval"
	"github.com/redesblock/mop/core/storer/netstore"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/storer/storage/inmemstore"
	"github.com/redesblock/mop/core/tracer"
)

var (
	// zeroed out while waiting to be  replacement for the new snapshot feed address
	// must be different to avoid stale reads on the old contract
	snapshotFeed    = cluster.MustParseHexAddress("0000000000000000000000000000000000000000000000000000000000000000")
	errDataMismatch = errors.New("data length mismatch")
)

const (
	getSnapshotRetries = 3
	retryWait          = time.Second * 5
	timeout            = time.Minute * 2
)

func bootstrapNode(
	addr string,
	clusterAddress cluster.Address,
	nonce []byte,
	chainID int64,
	overlayEthAddress common.Address,
	addressbook address.Interface,
	bootnodes []ma.Multiaddr,
	lightNodes *lightnode.Container,
	chequebookService chequebook.Service,
	chequeStore chequebook.ChequeStore,
	cashoutService chequebook.CashoutService,
	transactionService transaction.Service,
	stateStore storage.StateStorer,
	signer crypto.Signer,
	networkID uint64,
	logger log.Logger,
	libp2pPrivateKey *ecdsa.PrivateKey,
	o *Options,
) (snapshot *voucher.ChainSnapshot, retErr error) {

	tracer, tracerCloser, err := tracer.NewTracer(&tracer.Options{
		Enabled:     o.TracingEnabled,
		Endpoint:    o.TracingEndpoint,
		ServiceName: o.TracingServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("tracer: %w", err)
	}

	p2pCtx, p2pCancel := context.WithCancel(context.Background())

	b := &Mop{
		p2pCancel:    p2pCancel,
		tracerCloser: tracerCloser,
	}

	defer func() {
		retErr = multierror.Append(new(multierror.Error), retErr, b.Shutdown()).ErrorOrNil()
	}()

	p2ps, err := libp2p.New(p2pCtx, signer, networkID, clusterAddress, addr, addressbook, stateStore, lightNodes, logger, tracer, libp2p.Options{
		PrivateKey:     libp2pPrivateKey,
		NATAddr:        o.NATAddr,
		EnableWS:       o.EnableWS,
		WelcomeMessage: o.WelcomeMessage,
		FullNode:       false,
		Nonce:          nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("p2p service: %w", err)
	}
	b.p2pService = p2ps
	b.p2pHalter = p2ps

	hive, err := hive.New(p2ps, addressbook, networkID, o.BootnodeMode, o.AllowPrivateCIDRs, logger)
	if err != nil {
		return nil, fmt.Errorf("hive: %w", err)
	}

	if err = p2ps.AddProtocol(hive.Protocol()); err != nil {
		return nil, fmt.Errorf("hive service: %w", err)
	}
	b.hiveCloser = hive

	metricsDB, err := shed.NewDBWrap(stateStore.DB())
	if err != nil {
		return nil, fmt.Errorf("unable to create metrics storage for kademlia: %w", err)
	}

	kad, err := kademlia.New(clusterAddress, addressbook, hive, p2ps, &noopPinger{}, metricsDB, logger,
		kademlia.Options{Bootnodes: bootnodes, BootnodeMode: o.BootnodeMode, StaticNodes: o.StaticNodes})
	if err != nil {
		return nil, fmt.Errorf("unable to create kademlia: %w", err)
	}
	b.topologyCloser = kad
	b.topologyHalter = kad
	hive.SetAddPeersHandler(kad.AddPeers)
	p2ps.SetPickyNotifier(kad)

	paymentThreshold, _ := new(big.Int).SetString(o.PaymentThreshold, 10)
	lightPaymentThreshold := new(big.Int).Div(paymentThreshold, big.NewInt(lightFactor))

	pricer := pricer.NewFixedPricer(clusterAddress, basePrice)

	pricing := pricing.New(p2ps, logger, paymentThreshold, lightPaymentThreshold, big.NewInt(minPaymentThreshold))
	if err = p2ps.AddProtocol(pricing.Protocol()); err != nil {
		return nil, fmt.Errorf("pricing service: %w", err)
	}

	acc, err := bookkeeper.NewAccounting(
		paymentThreshold,
		o.PaymentTolerance,
		o.PaymentEarly,
		logger,
		stateStore,
		pricing,
		big.NewInt(refreshRate),
		lightFactor,
		p2ps,
	)
	if err != nil {
		return nil, fmt.Errorf("bookkeeper: %w", err)
	}
	b.accountingCloser = acc

	// bootstraper mode uses the light node refresh rate
	enforcedRefreshRate := big.NewInt(lightRefreshRate)

	pseudosettleService := pseudosettle.New(p2ps, logger, stateStore, acc, enforcedRefreshRate, enforcedRefreshRate, p2ps)
	if err = p2ps.AddProtocol(pseudosettleService.Protocol()); err != nil {
		return nil, fmt.Errorf("pseudosettle service: %w", err)
	}

	acc.SetRefreshFunc(pseudosettleService.Pay)

	pricing.SetPaymentThresholdObserver(acc)

	noopValidStamp := func(chunk cluster.Chunk, _ []byte) (cluster.Chunk, error) {
		return chunk, nil
	}

	storer := inmemstore.New()

	retrieve := retrieval.New(clusterAddress, storer, p2ps, kad, logger, acc, pricer, tracer, o.RetrievalCaching, noopValidStamp)
	if err = p2ps.AddProtocol(retrieve.Protocol()); err != nil {
		return nil, fmt.Errorf("retrieval service: %w", err)
	}

	ns := netstore.New(storer, noopValidStamp, retrieve, logger, o.MemCacheCapacity, o.TrustNode)

	if err := kad.Start(p2pCtx); err != nil {
		return nil, err
	}

	if err := p2ps.Ready(); err != nil {
		return nil, err
	}

	if !waitPeers(kad) {
		return nil, errors.New("timed out waiting for kademlia peers")
	}

	logger.Info("bootstrap: trying to fetch stamps snapshot")

	var (
		snapshotReference cluster.Address
		reader            file.Joiner
		l                 int64
		eventsJSON        []byte
	)

	for i := 0; i < getSnapshotRetries; i++ {
		if err != nil {
			time.Sleep(retryWait)
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		snapshotReference, err = getLatestSnapshot(ctx, ns, snapshotFeed)
		if err != nil {
			logger.Warning("bootstrap: fetching snapshot failed", "error", err)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}

	for i := 0; i < getSnapshotRetries; i++ {
		if err != nil {
			time.Sleep(retryWait)
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		reader, l, err = joiner.New(ctx, ns, snapshotReference)
		if err != nil {
			logger.Warning("bootstrap: file joiner failed", "error", err)
			continue
		}

		eventsJSON, err = io.ReadAll(reader)
		if err != nil {
			logger.Warning("bootstrap: reading failed", "error", err)
			continue
		}

		if len(eventsJSON) != int(l) {
			err = errDataMismatch
			logger.Warning("bootstrap: count mismatch", "error", err)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}

	events := voucher.ChainSnapshot{}
	err = json.Unmarshal(eventsJSON, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

// wait till some peers are connected. returns true if all is ok
func waitPeers(kad *kademlia.Kad) bool {
	for i := 0; i < 60; i++ {
		items := 0
		_ = kad.EachPeer(func(_ cluster.Address, _ uint8) (bool, bool, error) {
			items++
			return false, false, nil
		}, topology.Filter{})
		if items >= 25 {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

type noopPinger struct{}

func (p *noopPinger) Ping(context.Context, cluster.Address, ...string) (time.Duration, error) {
	return time.Duration(1), nil
}

func getLatestSnapshot(
	ctx context.Context,
	st storage.Storer,
	address cluster.Address,
) (cluster.Address, error) {
	ls := loadsave.NewReadonly(st)
	feedFactory := factory.New(st)

	m, err := manifest.NewDefaultManifestReference(
		address,
		ls,
	)
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("not a manifest: %w", err)
	}

	e, err := m.Lookup(ctx, "/")
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("node lookup: %w", err)
	}

	var (
		owner, topic []byte
		t            = new(feeds.Type)
	)
	meta := e.Metadata()
	if e := meta["cluster-feed-owner"]; e != "" {
		owner, err = hex.DecodeString(e)
		if err != nil {
			return cluster.ZeroAddress, err
		}
	}
	if e := meta["cluster-feed-topic"]; e != "" {
		topic, err = hex.DecodeString(e)
		if err != nil {
			return cluster.ZeroAddress, err
		}
	}
	if e := meta["cluster-feed-type"]; e != "" {
		err := t.FromString(e)
		if err != nil {
			return cluster.ZeroAddress, err
		}
	}
	if len(owner) == 0 || len(topic) == 0 {
		return cluster.ZeroAddress, fmt.Errorf("node lookup: %s", "feed metadata absent")
	}
	f := feeds.New(topic, common.BytesToAddress(owner))

	l, err := feedFactory.NewLookup(*t, f)
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("feed lookup failed: %w", err)
	}

	u, _, _, err := l.At(ctx, time.Now().Unix(), 0)
	if err != nil {
		return cluster.ZeroAddress, err
	}

	_, ref, err := feeds.FromChunk(u)
	if err != nil {
		return cluster.ZeroAddress, err
	}

	return cluster.NewAddress(ref), nil
}

func batchStoreExists(s storage.StateStorer) (bool, error) {

	hasOne := false
	err := s.Iterate("batchstore_", func(key, value []byte) (stop bool, err error) {
		hasOne = true
		return true, err
	})

	return hasOne, err
}
