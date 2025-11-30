package membership

import (
	"fmt"
	"io"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
)

// DebugList provides access to special debug functionality which should not be part of the List itself. These
// functionalities are helpful for writing tests and benchmarks.
// WARNING: This function MUST NOT be used in production!
func DebugList(list *List) *DebugListWrapper {
	return &DebugListWrapper{List: list}
}

// DebugListWrapper is the type which wraps the real List and exposes special functionality.
// WARNING: This struct MUST NOT be used in production!
type DebugListWrapper struct {
	*List
}

func (l *DebugListWrapper) WriteInternalDebugState(writer io.Writer) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, err := fmt.Fprintf(writer, "Membership List %s\n", l.self); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "Members (%d)\n", len(l.members)); err != nil {
		return err
	}
	for _, member := range l.members {
		if _, err := fmt.Fprintf(writer, "  - %s\n", member.Address); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(writer, "Next Direct Pings\n"); err != nil {
		return err
	}
	for i := l.nextRandomIndex; i < len(l.randomIndexes); i++ {
		if _, err := fmt.Fprintf(writer, "  - %s\n", l.members[l.randomIndexes[i]].Address); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(writer, "Gossip (%d)\n", l.gossipQueue.Len()); err != nil {
		return err
	}
	// Make sure we are not prioritizing for the state dump
	l.gossipQueue.Prioritize(encoding.Address{})
	l.gossipQueue.ForEach(func(message encoding.Message) bool {
		if _, err := fmt.Fprintf(writer, "  - %s\n", message); err != nil {
			return false
		}
		return true
	})
	return nil
}

func (l *DebugListWrapper) GetMembers() []encoding.Member {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.members
}

func (l *DebugListWrapper) GetFaultyMembers() []encoding.Member {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var result []encoding.Member
	l.faultyMembers.ForEach(func(member encoding.Member) bool {
		result = append(result, member)
		return true
	})
	return result
}

func (l *DebugListWrapper) SetMembers(members []encoding.Member) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.members = l.members[:0]
	l.randomIndexes = l.randomIndexes[:0]
	l.nextRandomIndex = 0

	for _, member := range members {
		l.addMember(member)
	}
}

func (l *DebugListWrapper) SetFaultyMembers(members []encoding.Member) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.faultyMembers.Clear()

	for _, member := range members {
		l.faultyMembers.Add(member)
	}
}

func (l *DebugListWrapper) GetGossip() *gossip.Queue {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.gossipQueue
}

func (l *DebugListWrapper) ClearGossip() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.gossipQueue.Clear()
}

// GetPendingDirectPings returns the current pending direct pings for testing.
func (d *DebugListWrapper) GetPendingDirectPings() []PendingDirectPing {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.pendingDirectPings
}

// GetPendingIndirectPings returns the current pending indirect pings for testing.
func (d *DebugListWrapper) GetPendingIndirectPings() []PendingIndirectPing {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.pendingIndirectPings
}

func (d *DebugListWrapper) SetConfig(config Config) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.config = config
}
