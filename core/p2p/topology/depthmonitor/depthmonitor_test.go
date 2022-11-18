package depthmonitor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	mockbatchstore "github.com/redesblock/mop/core/incentives/voucher/batchstore/mock"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p/topology"
	"github.com/redesblock/mop/core/p2p/topology/depthmonitor"
	"github.com/redesblock/mop/core/storer/storage"
)

func newTestSvc(
	t depthmonitor.Topology,
	s depthmonitor.SyncReporter,
	r depthmonitor.ReserveReporter,
	st storage.StateStorer,
	bs voucher.Storer,
	warmupTime time.Duration,
	wakeupInterval time.Duration,
) *depthmonitor.Service {

	var topo depthmonitor.Topology = &mockTopology{}
	if t != nil {
		topo = t
	}

	var syncer depthmonitor.SyncReporter = &mockSyncReporter{}
	if s != nil {
		syncer = s
	}

	var reserve depthmonitor.ReserveReporter = &mockReserveReporter{}
	if r != nil {
		reserve = r
	}

	batchStore := voucher.Storer(mockbatchstore.New())
	if bs != nil {
		batchStore = bs
	}

	return depthmonitor.New(topo, syncer, reserve, batchStore, log.Noop, warmupTime, wakeupInterval)
}

func TestDepthMonitorService(t *testing.T) {

	waitForDepth := func(t *testing.T, svc *depthmonitor.Service, depth uint8, timeout time.Duration) {
		t.Helper()
		start := time.Now()
		for {
			if time.Since(start) >= timeout {
				t.Fatalf("timed out waiting for depth expected %d found %d", depth, svc.StorageDepth())
			}
			if svc.StorageDepth() != depth {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			break
		}
	}

	t.Run("stop service within warmup time", func(t *testing.T) {
		svc := newTestSvc(nil, nil, nil, nil, nil, time.Second, depthmonitor.DefaultWakeupInterval)
		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("start with radius", func(t *testing.T) {
		bs := mockbatchstore.New(mockbatchstore.WithReserveState(&voucher.ReserveState{Radius: 3}))
		svc := newTestSvc(nil, nil, nil, nil, bs, 100*time.Millisecond, depthmonitor.DefaultWakeupInterval)
		waitForDepth(t, svc, 3, time.Second)
		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("depth decrease due to under utilization", func(t *testing.T) {
		const depthMonitorWakeUpInterval = 200 * time.Millisecond

		defer func(w uint8) {
			*depthmonitor.MinimumRadius = w
		}(*depthmonitor.MinimumRadius)
		*depthmonitor.MinimumRadius = 0

		topo := &mockTopology{peers: 1}
		// >50% utilized reserve
		reserve := &mockReserveReporter{size: 26000, capacity: 50000}

		bs := mockbatchstore.New(mockbatchstore.WithReserveState(&voucher.ReserveState{Radius: 3}))

		svc := newTestSvc(topo, nil, reserve, nil, bs, 100*time.Millisecond, depthMonitorWakeUpInterval)

		waitForDepth(t, svc, 3, time.Second)
		// simulate huge eviction to trigger manage worker
		reserve.setSize(1000)

		waitForDepth(t, svc, 1, time.Second)
		if topo.getStorageDepth() != 1 {
			t.Fatal("topology depth not updated")
		}
		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("depth doesnt change due to non-zero pull rate", func(t *testing.T) {
		const depthMonitorWakeUpInterval = 200 * time.Millisecond

		// under utilized reserve
		reserve := &mockReserveReporter{size: 10000, capacity: 50000}
		bs := mockbatchstore.New(mockbatchstore.WithReserveState(&voucher.ReserveState{Radius: 3}))
		syncer := &mockSyncReporter{rate: 10}

		svc := newTestSvc(nil, syncer, reserve, nil, bs, 100*time.Millisecond, depthMonitorWakeUpInterval)

		time.Sleep(2 * time.Second)
		// ensure that after few cycles of the adaptation period, the depth hasn't changed
		if svc.StorageDepth() != 3 {
			t.Fatal("found drop in depth")
		}
		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("depth doesnt change for utilized reserve", func(t *testing.T) {
		const depthMonitorWakeUpInterval = 200 * time.Millisecond

		// >50% utilized reserve
		reserve := &mockReserveReporter{size: 25001, capacity: 50000}
		bs := mockbatchstore.New(mockbatchstore.WithReserveState(&voucher.ReserveState{Radius: 3}))

		svc := newTestSvc(nil, nil, reserve, nil, bs, 100*time.Millisecond, depthMonitorWakeUpInterval)

		time.Sleep(2 * time.Second)
		// ensure the depth hasnt changed
		if svc.StorageDepth() != 3 {
			t.Fatal("found drop in depth")
		}
		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("radius setter handler", func(t *testing.T) {
		const depthMonitorWakeUpInterval = 200 * time.Millisecond

		topo := &mockTopology{connDepth: 3}
		bs := mockbatchstore.New(mockbatchstore.WithReserveState(&voucher.ReserveState{Radius: 3}))
		// >50% utilized reserve
		reserve := &mockReserveReporter{size: 25001, capacity: 50000}

		svc := newTestSvc(topo, nil, reserve, nil, bs, 100*time.Millisecond, depthMonitorWakeUpInterval)

		waitForDepth(t, svc, 3, time.Second)

		svc.SetStorageRadius(5)
		if svc.StorageDepth() != 5 {
			t.Fatalf("depth expected 5 found %d", svc.StorageDepth())
		}
		if topo.getStorageDepth() != 5 {
			t.Fatalf("topo depth expected 5 found %d", topo.getStorageDepth())
		}

		err := svc.Close()
		if err != nil {
			t.Fatal(err)
		}
	})
}

type mockTopology struct {
	sync.Mutex
	connDepth    uint8
	storageDepth uint8
	peers        int
}

func (m *mockTopology) NeighborhoodDepth() uint8 {
	return m.connDepth
}

func (m *mockTopology) IsWithinDepth(cluster.Address) bool {
	return false
}

func (m *mockTopology) SetStorageRadius(newDepth uint8) {
	m.Lock()
	defer m.Unlock()
	m.storageDepth = newDepth
}

func (m *mockTopology) PeersCount(topology.Filter) int {
	return m.peers
}

func (m *mockTopology) getStorageDepth() uint8 {
	m.Lock()
	defer m.Unlock()
	return m.storageDepth
}

type mockSyncReporter struct {
	rate float64
}

func (m *mockSyncReporter) Rate() float64 {
	return m.rate
}

type mockReserveReporter struct {
	sync.Mutex
	capacity uint64
	size     uint64
}

func (m *mockReserveReporter) ComputeReserveSize(uint8) (uint64, error) {
	m.Lock()
	defer m.Unlock()
	return m.size, nil
}

func (m *mockReserveReporter) setSize(sz uint64) {
	m.Lock()
	defer m.Unlock()
	m.size = sz
}

func (m *mockReserveReporter) ReserveCapacity() uint64 {
	return m.capacity
}
