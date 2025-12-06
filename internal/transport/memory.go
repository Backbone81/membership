package transport

import (
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/backbone81/membership/internal/encoding"
)

// Memory provides a transport implementation which moves network messages through memory between members. This is
// helpful for tests where you want to simulate the interaction between multiple members without having any real network
// involved. This transport can guarantee that network messages were delivered through its Flush calls. A normal network
// stack cannot provide such guarantees.
type Memory struct {
	mutex sync.Mutex

	// targets is the list of targets indexed by their address.
	targets map[encoding.Address]Target

	// pendingSends stores all sends for the given address for flushing later.
	pendingSends map[encoding.Address][][]byte

	// bufferPool is a pool of network buffers to avoid allocating too much new memory.
	bufferPool [][]byte
}

// NewMemory creates a new Memory instance.
func NewMemory() *Memory {
	return &Memory{
		targets:      make(map[encoding.Address]Target, 1024),
		pendingSends: make(map[encoding.Address][][]byte, 1024),
		bufferPool:   make([][]byte, 0, 1024),
	}
}

// Client returns a client transport for sending network messages through the memory transport.
func (m *Memory) Client() *MemoryClient {
	return &MemoryClient{
		memory: m,
	}
}

// AddTarget adds the target with the given address. The target will then receive network messages from clients.
func (m *Memory) AddTarget(address encoding.Address, target Target) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.targets[address] = target
}

// acquireBuffer is a helper function which either allocates a new memory buffer if none is available in the pool or
// if the buffer from the pool is not big enough.
func (m *Memory) acquireBuffer(length int) []byte {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.bufferPool) == 0 {
		return make([]byte, length)
	}

	buffer := m.bufferPool[len(m.bufferPool)-1]
	m.bufferPool = m.bufferPool[:len(m.bufferPool)-1]
	if cap(buffer) < length {
		return make([]byte, length)
	}

	return buffer[:length]
}

// AddPendingSend marks the buffer for being delivered later to the given address.
func (m *Memory) AddPendingSend(address encoding.Address, buffer []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.pendingSends[address] = append(m.pendingSends[address], buffer)
}

// FlushPendingSends dispatches all pending send to the target with the given address. If no target is registered for
// the given address, all pending sends are dropped.
func (m *Memory) FlushPendingSends(address encoding.Address) (bool, error) {
	// We are deliberately not locking the mutex here, because this might lead to a deadlock situation.
	// getPendingSends will acquire the lock on its own
	pendingSends := m.getPendingSends(address)
	if pendingSends == nil {
		return false, nil
	}

	// target will acquire the lock on its own
	target := m.target(address)
	if target == nil {
		return false, nil
	}
	var joinedError error
	for _, pendingSend := range pendingSends {
		if err := target.DispatchDatagram(pendingSend); err != nil {
			joinedError = errors.Join(joinedError, err)
		}
	}

	// Return back the buffers to the buffer pool for later re-user.
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.bufferPool = append(m.bufferPool, pendingSends...)

	return true, joinedError
}

// FlushAllPendingSends flushes all messages for all registered targets.
func (m *Memory) FlushAllPendingSends() error {
	var joinedError error
	repeatFlush := true
	for repeatFlush {
		repeatFlush = false
		for _, target := range m.Addresses() {
			sent, err := m.FlushPendingSends(target)
			if err != nil {
				joinedError = errors.Join(joinedError, err)
			}
			repeatFlush = repeatFlush || sent
		}
	}
	return joinedError
}

// Addresses returns a list with all addresses which are registered with this instance.
func (m *Memory) Addresses() []encoding.Address {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	addresses := slices.Collect(maps.Keys(m.targets))
	slices.SortFunc(addresses, encoding.CompareAddress)
	return addresses
}

// getPendingSends returns all byte buffers which are waiting to be sent to the given target. It will remove those
// buffers from the internal bookkeeping, which means that two consecutive calls to this function will not return the
// same result. Chances are the second call returns nil.
func (m *Memory) getPendingSends(address encoding.Address) [][]byte {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pendingSends, found := m.pendingSends[address]
	if !found {
		return nil
	}
	delete(m.pendingSends, address)
	return pendingSends
}

// target returns the target with the given address or nil.
func (m *Memory) target(address encoding.Address) Target {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	target, found := m.targets[address]
	if !found {
		return nil
	}
	return target
}

// MemoryClient is a client which used to communicate with other targets through the memory transport.
type MemoryClient struct {
	memory *Memory
}

// MemoryClient implements Transport.
var _ Transport = (*MemoryClient)(nil)

func (m *MemoryClient) Send(address encoding.Address, buffer []byte) error {
	copyBuffer := m.memory.acquireBuffer(len(buffer))
	copy(copyBuffer, buffer)
	m.memory.AddPendingSend(address, copyBuffer)
	return nil
}
