package membership

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
	"github.com/backbone81/membership/internal/utility"
)

// List provides the membership list as the implementation of the SWIM protocol.
//
// This data type should work independent of any temporal aspects to allow a clear separation between algorithm and
// scheduler. This means that this data type should have no knowledge about protocol period durations or timeouts. The
// algorithm is driven by calling DirectPing, IndirectPing and EndOfProtocolPeriod.
//
// List is safe for concurrent use by multiple goroutines. Access through exported methods is internally synchronized.
type List struct {
	// mutex is responsible for serializing concurrent access to members of this struct.
	mutex sync.Mutex

	// config holds the configuration of the membership list.
	config Config

	// logger provides the logger which the membership list uses to output status information.
	logger logr.Logger

	// self is the address for the current list instance. This is important because the own address might be hidden
	// behind a NAT router.
	self encoding.Address

	// nextSequenceNumber provides the sequence number to use for the next ping. Be aware that sequence number must only
	// ever be checked for equality and inequality. That way we do not have to deal with wrap-around events.
	nextSequenceNumber uint16

	// incarnationNumber is the incarnation number of this membership list instance. It is incremented each time this
	// list needs to refute a suspect or faulty message. Be aware that you need to use utility.IncarnationLessThan and
	// utility.IncarnationMax when dealing with incarnation numbers to correctly deal with wrap-around events.
	incarnationNumber uint16

	// members holds the list of members which are known to be alive or suspect. This list always needs to be sorted
	// by address to allow for binary searches in this list. It can contain thousands of elements in big clusters.
	members []encoding.Member

	// faultyMembers holds the list of members which were declared faulty. This is important for a full memberlist
	// sync to allow information about faulty members to be transported. This list always needs to be sorted by address
	// to allow for binary searches in this list. It can contain thousands of elements in big clusters.
	faultyMembers []encoding.Member

	// randomIndexes holds alist of random indexes into members. randomIndexes always has the same length as members and
	// every index only occurs once. This helps in having an upper bound on picking random members for direct pings.
	// If direct pings were truly random, there would be a slight chance that some member would never be picked as a
	// direct ping target by any other member. By shuffling all available members and working through that list before
	// shuffling again, we have a guarantee that in the works case after two sweeps through the member list each member
	// was picked at least once.
	randomIndexes []int

	// nextRandomIndex is the index into randomIndexes which describes the next random member to pick for direct pings.
	// This index is always increasing and wraps around at the end of randomIndexes - triggering a re-shuffle.
	nextRandomIndex int

	// gossipQueue provides the priority queue for gossip messages to piggyback on pings and acks.
	gossipQueue *gossip.Queue

	// datagramBuffer is the buffer to write network messages into. We re-use the same buffer for every network message
	// to reduce the amount of memory allocations happening. As access to this buffer is serialized on the top level,
	// we do not need more than one buffer as we cannot have more than one network message at the same time.
	datagramBuffer []byte

	// pendingDirectPings provides information about direct pings which were started in the current protocol period
	// and will end at the end of the current protocol period. These are direct pings which were triggered by the
	// scheduler as part of our own protocol cycle. This list will usually only contain a handful of elements and does
	// not require special ordering.
	pendingDirectPings []PendingDirectPing

	// pendingDirectPingsNext provides information about direct pings which were started in the current protocol period
	// and will end at the end of the NEXT protocol period. These are direct pings which were triggered in a response
	// to execute an indirect ping. As such requests can happen at any time, we need to keep them around until the end
	// of the next protocol period before we abandon them. This list will usually only contain a handful of elements and
	// does not require special ordering.
	pendingDirectPingsNext []PendingDirectPing

	// pendingIndirectPings provides information about indirect pings which were started in the current protocol period
	// and will end at the end of the current protocol period. These are indirect pings which were triggered by the
	// scheduler as part of our own protocol cycle. This list will usually only contain a handful of elements and does
	// not require special ordering.
	pendingIndirectPings []PendingIndirectPing
}

// NewList creates a new membership list.
func NewList(options ...Option) *List {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}

	if config.RoundTripTimeTracker == nil {
		panic("you must provide a round trip time tracker")
	}

	newList := List{
		config:                 config,
		logger:                 config.Logger,
		self:                   config.AdvertisedAddress,
		gossipQueue:            gossip.NewQueue(),
		datagramBuffer:         make([]byte, 0, config.MaxDatagramLengthSend),
		pendingDirectPings:     make([]PendingDirectPing, 0, 16),
		pendingDirectPingsNext: make([]PendingDirectPing, 0, 16),
		pendingIndirectPings:   make([]PendingIndirectPing, 0, 16),
	}

	// We need to gossip our own alive. Otherwise, nobody will pick us up into their own member list.
	newList.gossipQueue.Add(&gossip.MessageAlive{
		Source:            config.AdvertisedAddress,
		IncarnationNumber: 0,
	})
	for _, initialMember := range config.BootstrapMembers {
		newList.addMember(encoding.Member{
			Address:           initialMember,
			State:             encoding.MemberStateAlive,
			IncarnationNumber: 0,
		})
	}
	return &newList
}

// Config returns the configuration the membership list was created with.
func (l *List) Config() Config {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.config
}

// Len returns the number of members which are currently alive or suspect.
func (l *List) Len() int {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.members)
}

// Get returns the addresses of all members. The addresses are guaranteed to be sorted ascending. This call will always
// allocate a new slice holding the member addresses. Use GetInto if you want to avoid memory allocations.
func (l *List) Get() []encoding.Address {
	return l.GetInto(nil)
}

// GetInto returns the addresses of all members. The addresses are guaranteed to be sorted ascending. Providing a slice
// as parameter will use that slice to fill member addresses in. This allows the caller to prevent memory allocations.
// If the parameter is nil, a new slice will be allocated with capacity to hold the full member list.
func (l *List) GetInto(addresses []encoding.Address) []encoding.Address {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if addresses == nil {
		addresses = make([]encoding.Address, 0, len(l.members))
	} else {
		addresses = addresses[:0]
	}

	for _, member := range l.members {
		addresses = append(addresses, member.Address)
	}
	return addresses
}

// DirectPing executes the first step in the SWIM protocol by directly pinging other members.
func (l *List) DirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// As we are supporting to directly ping multiple members, we need to make sure that we are not exceeding the
	// current member count. Otherwise, we would ping the same member multiple times in the same protocol period.
	for range min(len(l.members), l.config.DirectPingMemberCount) {
		directPing := MessageDirectPing{
			Source:         l.self,
			SequenceNumber: l.nextSequenceNumber,
		}
		l.nextSequenceNumber++

		destination := l.getNextMember().Address
		l.logger.V(1).Info(
			"Direct ping",
			"source", l.self,
			"destination", destination,
			"sequence-number", directPing.SequenceNumber,
		)
		l.pendingDirectPings = append(l.pendingDirectPings, PendingDirectPing{
			Timestamp:         time.Now(),
			Destination:       destination,
			MessageDirectPing: directPing,
		})
		if err := l.sendWithGossip(destination, &directPing); err != nil {
			return err
		}
	}
	return nil
}

// getNextMember returns the next member to pick for a direct ping. This method ensures that the next member is selected
// randomly but with an upper bound to prevent some members never being picket for a direct ping. This is done by having
// a list of all indexes into the member list and shuffling that index list. Then we can move through that list one by
// one and re-shuffle once we reach the end of the list. This ensures that in the worst case a member is picked after
// two iterations of the member list.
func (l *List) getNextMember() *encoding.Member {
	if l.nextRandomIndex >= len(l.randomIndexes) {
		// When we moved beyond the end of the list, re-shuffle the indices and reset back to the start of the list.
		rand.Shuffle(len(l.randomIndexes), func(i, j int) {
			l.randomIndexes[i], l.randomIndexes[j] = l.randomIndexes[j], l.randomIndexes[i]
		})
		l.nextRandomIndex = 0
	}

	randomIndex := l.nextRandomIndex
	l.nextRandomIndex++
	return &l.members[l.randomIndexes[randomIndex]]
}

// IndirectPing executes the second step in the SWIM protocol by requesting indirect pings for direct pings which timed
// out.
func (l *List) IndirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// An indirect ping only makes sense whe we have at least two members.
	if len(l.members) < 2 {
		return nil
	}

	var joinedErr error
	for _, directPing := range l.pendingDirectPings {
		if !directPing.MessageIndirectPing.IsZero() {
			// We are not interested in direct pings which we do as a request for an indirect ping. We only do
			// indirect pings for direct pings we initiated on our own.
			continue
		}

		indirectPing := MessageIndirectPing{
			Source:      l.self,
			Destination: directPing.Destination,

			// We use the same sequence number as we did for the corresponding direct ping. That way, logs can
			// correlate the indirect ping with the direct ping, and we know which indirect pings to discard when
			// the direct ping succeeds late.
			SequenceNumber: directPing.MessageDirectPing.SequenceNumber,
		}

		// Send the indirect pings to the indirect ping members and join up all errors which might occur.
		members := l.pickIndirectPings(l.config.IndirectPingMemberCount, directPing.Destination)
		l.pendingIndirectPings = append(l.pendingIndirectPings, PendingIndirectPing{
			Timestamp:           time.Now(),
			MessageIndirectPing: indirectPing,
		})
		for _, member := range members {
			l.logger.V(1).Info(
				"Indirect ping",
				"source", l.self,
				"destination", indirectPing.Destination,
				"sequence-number", indirectPing.SequenceNumber,
				"through", member.Address,
			)
			if err := l.sendWithGossip(member.Address, &indirectPing); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		}
	}
	return joinedErr
}

func (l *List) pickIndirectPings(failureDetectionSubgroupSize int, directPingAddress encoding.Address) []*encoding.Member {
	// TODO: We should do a partial Fisher-Yates shuffle (https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle).
	// This is also what rand.Shuffle is doing for the full slice.
	candidateIndexes := make([]int, 0, len(l.members))
	for index := range l.members {
		if l.members[index].Address.Equal(directPingAddress) {
			// We do not want to include the direct ping member into our list for indirect pings.
			continue
		}
		candidateIndexes = append(candidateIndexes, index)
	}

	rand.Shuffle(len(candidateIndexes), func(i, j int) {
		candidateIndexes[i], candidateIndexes[j] = candidateIndexes[j], candidateIndexes[i]
	})

	// Pick the first few candidates.
	candidateIndexes = candidateIndexes[:min(failureDetectionSubgroupSize, len(candidateIndexes))]
	result := make([]*encoding.Member, len(candidateIndexes))
	for i := range candidateIndexes {
		result[i] = &l.members[candidateIndexes[i]]
	}
	return result
}

func (l *List) pickListRequestMember() *encoding.Member {
	members := l.pickIndirectPings(1, encoding.Address{})
	if len(members) < 1 {
		return nil
	}
	return members[0]
}

func (l *List) EndOfProtocolPeriod() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	maxTransmissionCount := l.requiredDisseminationPeriods()
	l.gossipQueue.SetMaxTransmissionCount(maxTransmissionCount)
	l.processFailedPings()
	l.markSuspectsAsFaulty()
	return nil
}

func (l *List) markSuspectsAsFaulty() {
	suspicionPeriodThreshold := l.requiredDisseminationPeriods()

	// As we are potentially removing elements from the member list, we need to iterate from the back to the front in
	// order to not skip a member when the content changes.
	for i := len(l.members) - 1; i >= 0; i-- {
		member := &l.members[i]
		if member.State != encoding.MemberStateSuspect {
			continue
		}
		member.SuspicionPeriodCounter++
		if member.SuspicionPeriodCounter < suspicionPeriodThreshold {
			continue
		}

		l.logger.Info(
			"Member declared as faulty",
			"source", l.self,
			"destination", member.Address,
			"incarnation-number", member.IncarnationNumber,
		)
		member.State = encoding.MemberStateFaulty
		faultyMemberIndex, found := slices.BinarySearchFunc(
			l.faultyMembers,
			*member,
			encoding.CompareMember,
		)
		if found {
			// Overwrite existing faulty member
			l.faultyMembers[faultyMemberIndex] = *member
		} else {
			l.faultyMembers = slices.Insert(l.faultyMembers, faultyMemberIndex, *member)
		}
		l.gossipQueue.Add(&gossip.MessageFaulty{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		})
		l.removeMemberByIndex(i) // must always happen last to keep the member alive during this method
	}
}

func (l *List) processFailedPings() {
	for _, pendingDirectPings := range l.pendingDirectPings {
		memberIndex, found := slices.BinarySearchFunc(
			l.members,
			encoding.Member{Address: pendingDirectPings.Destination},
			encoding.CompareMember,
		)
		if !found {
			// We probably got a faulty message by some other member while we were waiting for our ping to succeed.
			// Nothing to do here.
			continue
		}

		member := &l.members[memberIndex]
		if member.State == encoding.MemberStateSuspect {
			// The member is already suspect. Nothing to do.
			continue
		}

		l.logger.Info(
			"Member declared as suspect",
			"source", l.self,
			"destination", member.Address,
			"incarnation-number", member.IncarnationNumber,
		)

		// We need to mark the member as suspect and gossip about it.
		member.State = encoding.MemberStateSuspect
		member.SuspicionPeriodCounter = 0
		l.gossipQueue.Add(&gossip.MessageSuspect{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		})
	}
	l.pendingDirectPings, l.pendingDirectPingsNext = l.pendingDirectPingsNext, l.pendingDirectPings[:0]

	// As indirect pings always happen with a direct ping not being satisfied before, we can clear the indirect pings
	// without any further actions, as those actions have already been taken on the pending direct pings.
	l.pendingIndirectPings = l.pendingIndirectPings[:0]
}

func (l *List) RequestList() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	member := l.pickListRequestMember()
	if member == nil {
		// We could not pick a member to request the member list from. There probably is no other member at all.
		return nil
	}

	listRequest := MessageListRequest{
		Source: l.self,
	}

	l.logger.V(1).Info(
		"Requesting member list",
		"destination", member.Address,
	)
	if err := l.sendWithGossip(member.Address, &listRequest); err != nil {
		return err
	}
	return nil
}

func (l *List) DispatchDatagram(buffer []byte) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var joinedErr error
	for len(buffer) > 0 {
		messageType, _, err := encoding.MessageTypeFromBuffer(buffer)
		if err != nil {
			return err
		}

		switch messageType {
		case encoding.MessageTypeDirectPing:
			var message MessageDirectPing
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleDirectPing(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeDirectAck:
			var message MessageDirectAck
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleDirectAck(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeIndirectPing:
			var message MessageIndirectPing
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleIndirectPing(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeIndirectAck:
			var message MessageIndirectAck
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleIndirectAck(message)
		case encoding.MessageTypeSuspect:
			var message gossip.MessageSuspect
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleSuspect(message)
		case encoding.MessageTypeAlive:
			var message gossip.MessageAlive
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleAlive(message)
		case encoding.MessageTypeFaulty:
			var message gossip.MessageFaulty
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleFaulty(message)
		case encoding.MessageTypeListRequest:
			var message MessageListRequest
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleListRequest(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeListResponse:
			var message MessageListResponse
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleListResponse(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		default:
			log.Printf("ERROR: Unknown message type %d.", messageType)
		}
	}
	return joinedErr
}

func (l *List) handleDirectPing(directPing MessageDirectPing) error {
	l.logger.V(2).Info(
		"Received direct ping",
		"source", directPing.Source,
		"sequence-number", directPing.SequenceNumber,
	)
	directAck := MessageDirectAck{
		Source:         l.self,
		SequenceNumber: directPing.SequenceNumber,
	}
	if err := l.sendWithGossip(directPing.Source, &directAck); err != nil {
		return err
	}
	return nil
}

func (l *List) handleDirectAck(directAck MessageDirectAck) error {
	l.logger.V(2).Info(
		"Received direct ack",
		"source", directAck.Source,
		"sequence-number", directAck.SequenceNumber,
	)
	l.handleDirectAckForPendingIndirectPings(directAck)
	var err error
	l.pendingDirectPings, err = l.handleDirectAckForPendingDirectPings(l.pendingDirectPings, directAck)
	if err != nil {
		return err
	}
	l.pendingDirectPingsNext, err = l.handleDirectAckForPendingDirectPings(l.pendingDirectPingsNext, directAck)
	if err != nil {
		return err
	}
	return nil
}

func (l *List) handleDirectAckForPendingDirectPings(pendingDirectPings []PendingDirectPing, directAck MessageDirectAck) ([]PendingDirectPing, error) {
	// As we now got a direct ack, we don't have to wait for a direct ack anymore.
	pendingDirectPingIndex := slices.IndexFunc(pendingDirectPings, func(record PendingDirectPing) bool {
		return record.MessageDirectPing.SequenceNumber == directAck.SequenceNumber &&
			record.Destination.Equal(directAck.Source)
	})
	if pendingDirectPingIndex == -1 {
		return pendingDirectPings, nil
	}

	pendingDirectPing := pendingDirectPings[pendingDirectPingIndex]
	pendingDirectPings = utility.SwapDelete(pendingDirectPings, pendingDirectPingIndex)

	// We note down the round trip time for the direct ping.
	l.config.RoundTripTimeTracker.AddObserved(time.Since(pendingDirectPing.Timestamp))

	if pendingDirectPing.MessageIndirectPing.IsZero() {
		// The direct ping was NOT done in a response to a request for an indirect ping, so we are done here.
		return pendingDirectPings, nil
	}

	indirectAck := MessageIndirectAck{
		Source:         directAck.Source,
		SequenceNumber: pendingDirectPing.MessageIndirectPing.SequenceNumber,
	}
	if err := l.sendWithGossip(pendingDirectPing.MessageIndirectPing.Source, &indirectAck); err != nil {
		return pendingDirectPings, err
	}
	return pendingDirectPings, nil
}

func (l *List) handleDirectAckForPendingIndirectPings(directAck MessageDirectAck) {
	// We don't have to wait for the indirect ack anymore.
	pendingIndirectPingIndex := slices.IndexFunc(l.pendingIndirectPings, func(record PendingIndirectPing) bool {
		return record.MessageIndirectPing.SequenceNumber == directAck.SequenceNumber &&
			record.MessageIndirectPing.Destination.Equal(directAck.Source)
	})
	if pendingIndirectPingIndex == -1 {
		return
	}
	l.pendingIndirectPings = utility.SwapDelete(l.pendingIndirectPings, pendingIndirectPingIndex)
}

func (l *List) handleIndirectPing(indirectPing MessageIndirectPing) error {
	l.logger.V(2).Info(
		"Received indirect ping",
		"source", indirectPing.Source,
		"destination", indirectPing.Destination,
		"sequence-number", indirectPing.SequenceNumber,
	)
	directPing := MessageDirectPing{
		Source:         l.self,
		SequenceNumber: l.nextSequenceNumber,
	}
	l.nextSequenceNumber++

	l.pendingDirectPingsNext = append(l.pendingDirectPingsNext, PendingDirectPing{
		Timestamp:           time.Now(),
		Destination:         indirectPing.Destination,
		MessageDirectPing:   directPing,
		MessageIndirectPing: indirectPing,
	})

	if err := l.sendWithGossip(indirectPing.Destination, &directPing); err != nil {
		return err
	}
	return nil
}

func (l *List) handleIndirectAck(indirectAck MessageIndirectAck) {
	l.logger.V(2).Info(
		"Received indirect ack",
		"source", indirectAck.Source,
		"sequence-number", indirectAck.SequenceNumber,
	)
	l.handleIndirectAckForPendingDirectPings(indirectAck)
	l.handleIndirectAckForPendingIndirectPings(indirectAck)
}

func (l *List) handleIndirectAckForPendingDirectPings(indirectAck MessageIndirectAck) {
	// As we now got an indirect ack, we don't have to wait for a direct ack anymore.
	pendingDirectPingIndex := slices.IndexFunc(l.pendingDirectPings, func(record PendingDirectPing) bool {
		return record.MessageDirectPing.SequenceNumber == indirectAck.SequenceNumber &&
			record.Destination.Equal(indirectAck.Source)
	})
	if pendingDirectPingIndex == -1 {
		return
	}

	l.pendingDirectPings = utility.SwapDelete(l.pendingDirectPings, pendingDirectPingIndex)
}

func (l *List) handleIndirectAckForPendingIndirectPings(indirectAck MessageIndirectAck) {
	// We don't have to wait for the indirect ack anymore.
	pendingIndirectPingIndex := slices.IndexFunc(l.pendingIndirectPings, func(record PendingIndirectPing) bool {
		return record.MessageIndirectPing.SequenceNumber == indirectAck.SequenceNumber &&
			record.MessageIndirectPing.Destination.Equal(indirectAck.Source)
	})
	if pendingIndirectPingIndex == -1 {
		return
	}

	// We note down the round trip time for the indirect ping. Note that we divide by 2 because the indirect ping
	// consists of two round trips. We also add to observations, as we basically observed two round trips.
	observedRoundTrip := time.Since(l.pendingIndirectPings[pendingIndirectPingIndex].Timestamp) / 2
	l.config.RoundTripTimeTracker.AddObserved(observedRoundTrip)
	l.config.RoundTripTimeTracker.AddObserved(observedRoundTrip)

	l.pendingIndirectPings = utility.SwapDelete(l.pendingIndirectPings, pendingIndirectPingIndex)
}

func (l *List) handleSuspect(suspect gossip.MessageSuspect) {
	l.logger.V(3).Info(
		"Received gossip about suspect",
		"source", suspect.Source,
		"destination", suspect.Destination,
		"incarnation-number", suspect.IncarnationNumber,
	)
	if l.handleSuspectForSelf(suspect) {
		return
	}
	if l.handleSuspectForFaultyMembers(suspect) {
		return
	}
	if l.handleSuspectForMembers(suspect) {
		return
	}
	l.handleSuspectForUnknown(suspect)
}

func (l *List) handleSuspectForSelf(suspect gossip.MessageSuspect) bool {
	if !suspect.Destination.Equal(l.self) {
		return false
	}

	if utility.IncarnationLessThan(suspect.IncarnationNumber, l.incarnationNumber) {
		// We have a more up-to-date state than the gossip. Nothing to do.
		return true
	}

	// We need to refute the suspect about ourselves. Add a new alive message to gossip.
	// Also make sure that our incarnation number is bigger than before.
	l.incarnationNumber = utility.IncarnationMax(l.incarnationNumber+1, suspect.IncarnationNumber+1)
	l.gossipQueue.Add(&gossip.MessageAlive{
		Source:            l.self,
		IncarnationNumber: l.incarnationNumber,
	})
	return true
}

func (l *List) handleSuspectForFaultyMembers(suspect gossip.MessageSuspect) bool {
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		encoding.Member{Address: suspect.Destination},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if !utility.IncarnationLessThan(faultyMember.IncarnationNumber, suspect.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	// Move the faulty member over to the member list
	l.faultyMembers = slices.Delete(l.faultyMembers, faultyMemberIndex, faultyMemberIndex+1)
	l.addMember(encoding.Member{
		Address:                suspect.Destination,
		State:                  encoding.MemberStateSuspect,
		SuspicionPeriodCounter: 0,
		IncarnationNumber:      suspect.IncarnationNumber,
	})
	l.gossipQueue.Add(&suspect)
	return true
}

func (l *List) handleSuspectForMembers(suspect gossip.MessageSuspect) bool {
	memberIndex, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: suspect.Destination},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our members list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if utility.IncarnationLessThan(suspect.IncarnationNumber, member.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	member.IncarnationNumber = suspect.IncarnationNumber
	if member.State == encoding.MemberStateSuspect {
		// We already know about this member being suspect. Nothing to do.
		return true
	}

	// This information is new to us, we need to make sure to gossip about it.
	member.State = encoding.MemberStateSuspect
	member.SuspicionPeriodCounter = 0
	l.gossipQueue.Add(&suspect)
	return true
}

func (l *List) handleSuspectForUnknown(suspect gossip.MessageSuspect) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:                suspect.Destination,
		State:                  encoding.MemberStateSuspect,
		SuspicionPeriodCounter: 0,
		IncarnationNumber:      suspect.IncarnationNumber,
	})
	l.gossipQueue.Add(&suspect)
}

func (l *List) handleAlive(alive gossip.MessageAlive) {
	l.logger.V(3).Info(
		"Received gossip about alive",
		"source", alive.Source,
		"incarnation-number", alive.IncarnationNumber,
	)
	if l.handleAliveForSelf(alive) {
		return
	}
	if l.handleAliveForFaultyMembers(alive) {
		return
	}
	if l.handleAliveForMembers(alive) {
		return
	}
	l.handleAliveForUnknown(alive)
}

func (l *List) handleAliveForSelf(alive gossip.MessageAlive) bool {
	return alive.Source.Equal(l.self)
}

func (l *List) handleAliveForFaultyMembers(alive gossip.MessageAlive) bool {
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		encoding.Member{Address: alive.Source},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if !utility.IncarnationLessThan(faultyMember.IncarnationNumber, alive.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	// Move the faulty member over to the member list
	l.faultyMembers = slices.Delete(l.faultyMembers, faultyMemberIndex, faultyMemberIndex+1)
	l.addMember(encoding.Member{
		Address:           alive.Source,
		State:             encoding.MemberStateAlive,
		IncarnationNumber: alive.IncarnationNumber,
	})
	l.gossipQueue.Add(&alive)
	return true
}

func (l *List) handleAliveForMembers(alive gossip.MessageAlive) bool {
	memberIndex, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: alive.Source},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our members list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if !utility.IncarnationLessThan(member.IncarnationNumber, alive.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	member.IncarnationNumber = alive.IncarnationNumber
	if member.State == encoding.MemberStateAlive {
		// We already know about this member being alive. Nothing to do.
		return true
	}

	// This information is new to us, we need to make sure to gossip about it.
	member.State = encoding.MemberStateAlive
	l.gossipQueue.Add(&alive)
	return true
}

func (l *List) handleAliveForUnknown(alive gossip.MessageAlive) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:           alive.Source,
		State:             encoding.MemberStateAlive,
		IncarnationNumber: alive.IncarnationNumber,
	})
	l.gossipQueue.Add(&alive)
}

func (l *List) handleFaulty(faulty gossip.MessageFaulty) {
	l.logger.V(3).Info(
		"Received gossip about faulty",
		"source", faulty.Source,
		"destination", faulty.Destination,
		"incarnation-number", faulty.IncarnationNumber,
	)
	if l.handleFaultyForSelf(faulty) {
		return
	}
	if l.handleFaultyForFaultyMembers(faulty) {
		return
	}
	if l.handleFaultyForMembers(faulty) {
		return
	}
	l.handleFaultyForUnknown(faulty)
}

func (l *List) handleFaultyForSelf(faulty gossip.MessageFaulty) bool {
	if !faulty.Destination.Equal(l.self) {
		return false
	}

	if utility.IncarnationLessThan(faulty.IncarnationNumber, l.incarnationNumber) {
		// We have a more up-to-date state than the gossip. Nothing to do.
		return true
	}

	// We need to re-join. Add a new alive message to gossip.
	// Also make sure that our incarnation number is bigger than before.
	l.incarnationNumber = utility.IncarnationMax(l.incarnationNumber+1, faulty.IncarnationNumber+1)
	l.gossipQueue.Add(&gossip.MessageAlive{
		Source:            l.self,
		IncarnationNumber: l.incarnationNumber,
	})
	return true
}

func (l *List) handleFaultyForFaultyMembers(faulty gossip.MessageFaulty) bool {
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		encoding.Member{Address: faulty.Destination},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if utility.IncarnationLessThan(faulty.IncarnationNumber, faultyMember.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	// Update the incarnation number to make sure we have the most current incarnation.
	faultyMember.IncarnationNumber = faulty.IncarnationNumber
	return true
}

func (l *List) handleFaultyForMembers(faulty gossip.MessageFaulty) bool {
	memberIndex, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: faulty.Destination},
		encoding.CompareMember,
	)
	if !found {
		// The member is not part of our member list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if utility.IncarnationLessThan(faulty.IncarnationNumber, member.IncarnationNumber) {
		// We have more up-to-date information about this member.
		return true
	}

	// Remove member from member list and put it on the faulty member list.
	member.State = encoding.MemberStateFaulty
	member.IncarnationNumber = faulty.IncarnationNumber
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		*member,
		encoding.CompareMember,
	)
	if found {
		// Overwrite existing faulty member
		l.faultyMembers[faultyMemberIndex] = *member
	} else {
		l.faultyMembers = slices.Insert(l.faultyMembers, faultyMemberIndex, *member)
	}
	l.gossipQueue.Add(&faulty)
	l.removeMemberByAddress(faulty.Destination) // must always happen last to keep the member alive during this method
	return true
}

func (l *List) handleFaultyForUnknown(faulty gossip.MessageFaulty) {
	// We don't know about this member yet. Add it to our faulty member list and gossip about it.
	faultyMember := encoding.Member{
		Address:           faulty.Destination,
		State:             encoding.MemberStateFaulty,
		IncarnationNumber: faulty.IncarnationNumber,
	}
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		faultyMember,
		encoding.CompareMember,
	)
	if found {
		// Overwrite existing faulty member
		l.faultyMembers[faultyMemberIndex] = faultyMember
	} else {
		l.faultyMembers = slices.Insert(l.faultyMembers, faultyMemberIndex, faultyMember)
	}
	l.gossipQueue.Add(&faulty)
}

func (l *List) handleListRequest(listRequest MessageListRequest) error {
	l.logger.V(2).Info(
		"Received list request",
		"source", listRequest.Source,
	)

	listResponse := MessageListResponse{
		Source:  l.self,
		Members: append(l.members, l.faultyMembers...),
	}
	buffer, _, err := listResponse.AppendToBuffer(l.datagramBuffer[:0])
	if err != nil {
		return err
	}

	if err := l.config.TCPClient.Send(listRequest.Source, buffer); err != nil {
		return err
	}
	return nil
}

func (l *List) handleListResponse(listResponse MessageListResponse) error {
	l.logger.V(2).Info(
		"Received list response",
		"source", listResponse.Source,
	)
	for _, member := range listResponse.Members {
		switch member.State {
		case encoding.MemberStateAlive:
			l.handleAlive(gossip.MessageAlive{
				Source:            listResponse.Source,
				IncarnationNumber: member.IncarnationNumber,
			})
		case encoding.MemberStateSuspect:
			l.handleSuspect(gossip.MessageSuspect{
				Source:            listResponse.Source,
				Destination:       member.Address,
				IncarnationNumber: member.IncarnationNumber,
			})
		case encoding.MemberStateFaulty:
			l.handleFaulty(gossip.MessageFaulty{
				Source:            listResponse.Source,
				Destination:       member.Address,
				IncarnationNumber: member.IncarnationNumber,
			})
		default:
			return fmt.Errorf("unknown member state: %v", member.State)
		}
	}
	return nil
}

func (l *List) sendWithGossip(address encoding.Address, message Message) error {
	l.datagramBuffer = l.datagramBuffer[:0]

	var err error
	var datagramN int
	l.datagramBuffer, datagramN, err = message.AppendToBuffer(l.datagramBuffer)
	if err != nil {
		return err
	}

	l.gossipQueue.Prioritize(address)
	var gossipAdded int
	for i := range l.gossipQueue.Len() {
		var gossipN int
		l.datagramBuffer, gossipN, err = l.gossipQueue.Get(i).AppendToBuffer(l.datagramBuffer)
		if err != nil {
			return err
		}

		if len(l.datagramBuffer) > l.config.MaxDatagramLengthSend {
			l.datagramBuffer = l.datagramBuffer[:len(l.datagramBuffer)-gossipN]
			break
		}

		datagramN += gossipN
		gossipAdded++
	}
	l.gossipQueue.MarkTransmitted(gossipAdded)

	if err := l.config.UDPClient.Send(address, l.datagramBuffer); err != nil {
		return err
	}
	return nil
}

func (l *List) addMember(member encoding.Member) {
	if member.Address.Equal(l.self) {
		// We do not add ourselves to the member list
		return
	}

	l.logger.Info(
		"Member added",
		"address", member.Address,
		"incarnation-number", member.IncarnationNumber,
	)

	// Insert the new member at the correct ordered location. Remember the index for later.
	memberIndex, found := slices.BinarySearchFunc(
		l.members,
		member,
		encoding.CompareMember,
	)
	if found {
		// Update the existing member. Note that we do not count this towards the add member metric. Otherwise, the
		// number of members could not be calculated by subtracting remove member metric from add member metric.
		l.members[memberIndex] = member
		return
	}
	l.members = slices.Insert(l.members, memberIndex, member)

	// Fix the current indices to account for the inserted member.
	for i := range l.randomIndexes {
		if l.randomIndexes[i] < memberIndex {
			continue
		}
		l.randomIndexes[i]++
	}

	// We pick a random location to insert the new member into the random indexes slice. We need to add +1 to the length
	// of that slice to allow for appending at the end.
	insertIndex := rand.Intn(len(l.randomIndexes) + 1)
	l.randomIndexes = slices.Insert(l.randomIndexes, insertIndex, memberIndex)
	if insertIndex <= l.nextRandomIndex {
		// The new member index was inserted before or at the next random index. We therefore move the next random index
		// forward by one to not have the same member be picked twice in a row.
		l.nextRandomIndex++
	}

	// Trigger the callback if set.
	if l.config.MemberAddedCallback != nil {
		l.config.MemberAddedCallback(member.Address)
	}
	MembersAddedTotal.Inc()
}

func (l *List) removeMemberByAddress(address encoding.Address) {
	index, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: address},
		encoding.CompareMember,
	)
	if !found {
		return
	}
	l.removeMemberByIndex(index) // must always happen last to keep the member alive during this method
}

func (l *List) removeMemberByIndex(index int) {
	l.logger.Info(
		"Member removed",
		"address", l.members[index].Address,
		"incarnation-number", l.members[index].IncarnationNumber,
	)

	// Trigger the callback if set.
	if l.config.MemberRemovedCallback != nil {
		l.config.MemberRemovedCallback(l.members[index].Address)
	}

	// If we remove the element from the slice, all indexes after that element need to be decremented by one.
	var randomIndex int
	for i := range l.randomIndexes {
		if index < l.randomIndexes[i] {
			l.randomIndexes[i]--
		}
		if index == l.randomIndexes[i] {
			randomIndex = i
		}
	}

	l.members = slices.Delete(l.members, index, index+1)
	l.randomIndexes = slices.Delete(l.randomIndexes, randomIndex, randomIndex+1)
	MembersRemovedTotal.Inc()
}

func (l *List) WriteInternalDebugState(writer io.Writer) error {
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
	for i := range l.gossipQueue.Len() {
		gossipMessage := l.gossipQueue.Get(i)
		if _, err := fmt.Fprintf(writer, "  - %s\n", gossipMessage); err != nil {
			return err
		}
	}
	return nil
}

func (l *List) requiredDisseminationPeriods() int {
	return int(math.Ceil(utility.DisseminationPeriods(l.config.SafetyFactor, len(l.members))))
}

func (l *List) BroadcastShutdown() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	faultyMessage := gossip.MessageFaulty{
		Source:            l.self,
		Destination:       l.self,
		IncarnationNumber: l.incarnationNumber,
	}

	members := l.pickShutdownMembers(l.config.ShutdownMemberCount)
	for _, member := range members {
		l.logger.Info(
			"Broadcasting shutdown",
			"address", member.Address,
		)
		// We send our broadcast with the gossip we have, to help disseminate that information before we are gone.
		if err := l.sendWithGossip(member.Address, &faultyMessage); err != nil {
			return err
		}
	}
	return nil
}

func (l *List) pickShutdownMembers(memberCount int) []*encoding.Member {
	candidateIndexes := make([]int, 0, len(l.members))
	for index := range l.members {
		candidateIndexes = append(candidateIndexes, index)
	}

	rand.Shuffle(len(candidateIndexes), func(i, j int) {
		candidateIndexes[i], candidateIndexes[j] = candidateIndexes[j], candidateIndexes[i]
	})

	// Pick the first few candidates.
	candidateIndexes = candidateIndexes[:min(memberCount, len(candidateIndexes))]
	result := make([]*encoding.Member, len(candidateIndexes))
	for i := range candidateIndexes {
		result[i] = &l.members[candidateIndexes[i]]
	}
	return result
}

func (l *List) GetMembers() []encoding.Member {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.members
}

func (l *List) GetFaultyMembers() []encoding.Member {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.faultyMembers
}

func (l *List) SetMembers(members []encoding.Member) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.members = l.members[:0]
	l.randomIndexes = l.randomIndexes[:0]
	l.nextRandomIndex = 0

	for _, member := range members {
		l.addMember(member)
	}
}

func (l *List) SetFaultyMembers(members []encoding.Member) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.faultyMembers = append(l.faultyMembers[:0], members...)
}

func (l *List) GetGossip() *gossip.Queue {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.gossipQueue
}

func (l *List) ClearGossip() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.gossipQueue.Clear()
}
