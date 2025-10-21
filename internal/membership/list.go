package membership

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
	"github.com/go-logr/logr"
)

type List struct {
	mutex sync.Mutex

	config Config
	logger logr.Logger

	nextSequenceNumber           int
	incarnationNumber            int
	failureDetectionSubgroupSize int

	members       []encoding.Member
	faultyMembers []encoding.Member

	randomIndexes   []int
	nextRandomIndex int
	gossipQueue     *gossip.MessageQueue
	self            encoding.Address

	datagramBuffer []byte

	pendingDirectProbes   []DirectProbeRecord
	pendingIndirectProbes []IndirectProbeRecord
}

// DirectProbeRecord provides bookkeeping for a direct probe which is still active.
type DirectProbeRecord struct {
	// Timestamp is the point in time the direct probe was initiated.
	Timestamp time.Time

	// Destination is the address which the direct probe was sent to.
	Destination encoding.Address

	// MessageDirectPing is a copy of the message which was sent for the direct probe.
	MessageDirectPing MessageDirectPing

	// MessageIndirectPing is a copy of a received indirect probe request. It is the zero value in case the direct
	// probe was not initiated in response to an indirect probe request.
	MessageIndirectPing MessageIndirectPing
}

// IndirectProbeRecord provides bookkeeping for an indirect probe which is still active.
type IndirectProbeRecord struct {
	// Timestamp is the point in time the indirect probe was initiated.
	Timestamp time.Time

	// MessageIndirectPing is a copy of the message which was sent for an indirect probe.
	MessageIndirectPing MessageIndirectPing
}

func NewList(options ...Option) *List {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}

	newList := List{
		config:         config,
		logger:         config.Logger,
		self:           config.AdvertisedAddress,
		gossipQueue:    gossip.NewMessageQueue(10), // TODO: The max gossip count needs to be adjusted for the number of members during runtime.
		datagramBuffer: make([]byte, 0, config.MaxDatagramLength),
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
			LastStateChange:   time.Now(),
			IncarnationNumber: 0,
		})
	}
	return &newList
}

func (l *List) Len() int {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.members)
}

func (l *List) Get() []encoding.Address {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	result := make([]encoding.Address, 0, len(l.members))
	for _, member := range l.members {
		result = append(result, member.Address)
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

func (l *List) DirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.members) < 1 {
		// As long as there are no other members, we don't have to probe anyone.
		return nil
	}

	directPing := MessageDirectPing{
		Source:         l.self,
		SequenceNumber: l.nextSequenceNumber,
	}
	l.nextSequenceNumber = (l.nextSequenceNumber + 1) % math.MaxUint16

	destination := l.getNextMember().Address
	l.logger.V(1).Info(
		"Direct probe",
		"source", l.self,
		"destination", destination,
		"sequence-number", directPing.SequenceNumber,
	)
	l.pendingDirectProbes = append(l.pendingDirectProbes, DirectProbeRecord{
		Timestamp:         time.Now(),
		Destination:       destination,
		MessageDirectPing: directPing,
	})
	if err := l.sendWithGossip(destination, &directPing); err != nil {
		return err
	}
	return nil
}

func (l *List) IndirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// An indirect probe only makes sense whe we have at least two members.
	if len(l.members) < 2 {
		return nil
	}

	var joinedErr error
	for _, directProbe := range l.pendingDirectProbes {
		if !directProbe.MessageIndirectPing.IsZero() {
			// We are not interested in direct probes which we do as a request for an indirect probe. We only do
			// indirect probes for direct probes we initiated on our own.
			continue
		}

		indirectPing := MessageIndirectPing{
			Source:      l.self,
			Destination: directProbe.Destination,

			// We use the same sequence number as we did for the corresponding direct probe. That way, logs can
			// correlate the indirect probe with the direct probe, and we know which indirect probes to discard when
			// the direct probe succeeds late.
			SequenceNumber: directProbe.MessageDirectPing.SequenceNumber,
		}

		// Send the indirect probes to the indirect probe members and join up all errors which might occur.
		members := l.pickIndirectProbes(l.failureDetectionSubgroupSize, directProbe.Destination)
		l.pendingIndirectProbes = append(l.pendingIndirectProbes, IndirectProbeRecord{
			Timestamp:           time.Now(),
			MessageIndirectPing: indirectPing,
		})
		for _, member := range members {
			l.logger.V(1).Info(
				"Indirect probe",
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

func (l *List) pickIndirectProbes(failureDetectionSubgroupSize int, directProbeAddress encoding.Address) []*encoding.Member {
	candidateIndexes := make([]int, 0, len(l.members))
	for index := range l.members {
		if l.members[index].Address.Equal(directProbeAddress) {
			// We do not want to include the direct probe member into our list for indirect probes.
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

func (l *List) EndOfProtocolPeriod() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.markSuspectsAsFaulty()
	l.processFailedProbes()
	return nil
}

func (l *List) markSuspectsAsFaulty() {
	suspicionThreshold := time.Now().Add(-SuspicionTimeout(l.config.ProtocolPeriod, 3, len(l.members)))

	// As we are potentially removing elements from the member list, we need to iterate from the back to the front in
	// order to not skip a member when the content changes.
	for i := len(l.members) - 1; i >= 0; i-- {
		member := &l.members[i]
		if member.State != encoding.MemberStateSuspect {
			continue
		}
		if !member.LastStateChange.Before(suspicionThreshold) {
			continue
		}

		l.logger.Info(
			"Member declared as faulty",
			"source", l.self,
			"destination", member.Address,
			"incarnation-number", member.IncarnationNumber,
		)
		l.faultyMembers = append(l.faultyMembers, encoding.Member{
			Address:           member.Address,
			State:             encoding.MemberStateFaulty,
			LastStateChange:   time.Now(),
			IncarnationNumber: member.IncarnationNumber,
		})
		l.gossipQueue.Add(&gossip.MessageFaulty{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		})
		l.removeMemberByIndex(i)
	}
}

func (l *List) processFailedProbes() {
	timeout := time.Now().Add(-l.config.ProtocolPeriod)
	for _, pendingDirectProbe := range l.pendingDirectProbes {
		if !pendingDirectProbe.MessageIndirectPing.IsZero() && !pendingDirectProbe.Timestamp.Before(timeout) {
			// This is a direct probe which we need to keep around. Those are always requests for indirect probes.
			continue
		}

		memberIndex := slices.IndexFunc(l.members, func(record encoding.Member) bool {
			return record.Address.Equal(pendingDirectProbe.Destination)
		})
		if memberIndex == -1 {
			// We probably got a faulty message by some other member while we were waiting for our probe to succeed.
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
		member.LastStateChange = time.Now()
		l.gossipQueue.Add(&gossip.MessageSuspect{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		})
	}

	// Remove all pending probes which have timed out.
	l.pendingDirectProbes = slices.DeleteFunc(l.pendingDirectProbes, func(record DirectProbeRecord) bool {
		return record.MessageIndirectPing.IsZero() || record.Timestamp.Before(timeout)
	})

	// As indirect probes always happen with a direct probe not being satisfied before, we can clear the indirect probes
	// without any further actions, as those actions have already been taken on the pending direct probes.
	l.pendingIndirectProbes = l.pendingIndirectProbes[:0]
}

func (l *List) RequestList() error {
	// Disabled for now.
	return nil

	//if time.Since(l.lastListRequest) <= l.listRequestInterval {
	//	// No need to request another member list, as we already have a quite up-to-date list from somebody else.
	//	return nil
	//}
	//
	//l.lastListRequest = time.Now()
	//listRequest := MessageListRequest{
	//	Source: l.self,
	//}
	//
	//members := l.pickIndirectProbes(1, Address{})
	//var joinedErr error
	//for _, member := range members {
	//	l.logger.V(1).Info(
	//		"Requesting member list",
	//		"destination", member.Address,
	//	)
	//	if err := l.sendWithGossip(member.Address, &listRequest); err != nil {
	//		joinedErr = errors.Join(joinedErr, err)
	//	}
	//}
	//return joinedErr
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
	l.handleDirectAckForPendingIndirectProbes(directAck)
	if err := l.handleDirectAckForPendingDirectProbes(directAck); err != nil {
		return err
	}
	return nil
}

func (l *List) handleDirectAckForPendingDirectProbes(directAck MessageDirectAck) error {
	// As we now got an indirect ack, we don't have to wait for a direct ack anymore.
	pendingDirectProbeIndex := slices.IndexFunc(l.pendingDirectProbes, func(record DirectProbeRecord) bool {
		return record.MessageDirectPing.SequenceNumber == directAck.SequenceNumber &&
			record.Destination.Equal(directAck.Source)
	})
	if pendingDirectProbeIndex == -1 {
		return nil
	}

	pendingDirectProbe := l.pendingDirectProbes[pendingDirectProbeIndex]
	l.pendingDirectProbes = slices.Delete(l.pendingDirectProbes, pendingDirectProbeIndex, pendingDirectProbeIndex+1)

	if pendingDirectProbe.MessageIndirectPing.IsZero() {
		// The direct probe was NOT done in a response to a request for an indirect probe, so we are done here.
		return nil
	}

	indirectAck := MessageIndirectAck{
		Source:         directAck.Source,
		SequenceNumber: pendingDirectProbe.MessageIndirectPing.SequenceNumber,
	}
	if err := l.sendWithGossip(pendingDirectProbe.MessageIndirectPing.Source, &indirectAck); err != nil {
		return err
	}
	return nil
}

func (l *List) handleDirectAckForPendingIndirectProbes(directAck MessageDirectAck) {
	// We don't have to wait for the indirect ack anymore.
	pendingIndirectProbeIndex := slices.IndexFunc(l.pendingIndirectProbes, func(record IndirectProbeRecord) bool {
		return record.MessageIndirectPing.SequenceNumber == directAck.SequenceNumber &&
			record.MessageIndirectPing.Destination.Equal(directAck.Source)
	})
	if pendingIndirectProbeIndex == -1 {
		return
	}
	l.pendingIndirectProbes = slices.Delete(l.pendingIndirectProbes, pendingIndirectProbeIndex, pendingIndirectProbeIndex+1)
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
	l.nextSequenceNumber = (l.nextSequenceNumber + 1) % math.MaxUint16

	l.pendingDirectProbes = append(l.pendingDirectProbes, DirectProbeRecord{
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
	l.handleIndirectAckForPendingDirectProbes(indirectAck)
	l.handleIndirectAckForPendingIndirectProbes(indirectAck)
}

func (l *List) handleIndirectAckForPendingDirectProbes(indirectAck MessageIndirectAck) {
	// As we now got an indirect ack, we don't have to wait for a direct ack anymore.
	pendingDirectProbeIndex := slices.IndexFunc(l.pendingDirectProbes, func(record DirectProbeRecord) bool {
		return record.MessageDirectPing.SequenceNumber == indirectAck.SequenceNumber &&
			record.Destination.Equal(indirectAck.Source)
	})
	if pendingDirectProbeIndex == -1 {
		return
	}

	l.pendingDirectProbes = slices.Delete(l.pendingDirectProbes, pendingDirectProbeIndex, pendingDirectProbeIndex+1)
}

func (l *List) handleIndirectAckForPendingIndirectProbes(indirectAck MessageIndirectAck) {
	// We don't have to wait for the indirect ack anymore.
	pendingIndirectProbeIndex := slices.IndexFunc(l.pendingIndirectProbes, func(record IndirectProbeRecord) bool {
		return record.MessageIndirectPing.SequenceNumber == indirectAck.SequenceNumber &&
			record.MessageIndirectPing.Destination.Equal(indirectAck.Source)
	})
	if pendingIndirectProbeIndex == -1 {
		return
	}
	l.pendingIndirectProbes = slices.Delete(l.pendingIndirectProbes, pendingIndirectProbeIndex, pendingIndirectProbeIndex+1)
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

	if suspect.IncarnationNumber < l.incarnationNumber {
		// We have a more up-to-date state than the gossip. Nothing to do.
		return true
	}

	// We need to refute the suspect about ourselves. Add a new alive message to gossip.
	// Also make sure that our incarnation number is bigger than before.
	l.incarnationNumber = max(l.incarnationNumber+1, suspect.IncarnationNumber+1)
	l.gossipQueue.Add(&gossip.MessageAlive{
		Source:            l.self,
		IncarnationNumber: l.incarnationNumber,
	})
	return true
}

func (l *List) handleSuspectForFaultyMembers(suspect gossip.MessageSuspect) bool {
	faultyMemberIndex := slices.IndexFunc(l.faultyMembers, func(member encoding.Member) bool {
		return member.Address.Equal(suspect.Destination)
	})
	if faultyMemberIndex == -1 {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if suspect.IncarnationNumber <= faultyMember.IncarnationNumber {
		// We have more up-to-date information about this member.
		return true
	}

	// Move the faulty member over to the member list
	l.faultyMembers = slices.Delete(l.faultyMembers, faultyMemberIndex, faultyMemberIndex+1)
	l.addMember(encoding.Member{
		Address:           suspect.Destination,
		State:             encoding.MemberStateSuspect,
		LastStateChange:   time.Now(),
		IncarnationNumber: suspect.IncarnationNumber,
	})
	l.gossipQueue.Add(&suspect)
	return true
}

func (l *List) handleSuspectForMembers(suspect gossip.MessageSuspect) bool {
	memberIndex := slices.IndexFunc(l.members, func(member encoding.Member) bool {
		return member.Address.Equal(suspect.Destination)
	})
	if memberIndex == -1 {
		// The member is not part of our members list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if suspect.IncarnationNumber < member.IncarnationNumber {
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
	member.LastStateChange = time.Now()
	l.gossipQueue.Add(&suspect)
	return true
}

func (l *List) handleSuspectForUnknown(suspect gossip.MessageSuspect) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:           suspect.Destination,
		State:             encoding.MemberStateSuspect,
		LastStateChange:   time.Now(),
		IncarnationNumber: suspect.IncarnationNumber,
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
	if !alive.Source.Equal(l.self) {
		return false
	}
	return true
}

func (l *List) handleAliveForFaultyMembers(alive gossip.MessageAlive) bool {
	faultyMemberIndex := slices.IndexFunc(l.faultyMembers, func(member encoding.Member) bool {
		return member.Address.Equal(alive.Source)
	})
	if faultyMemberIndex == -1 {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if alive.IncarnationNumber <= faultyMember.IncarnationNumber {
		// We have more up-to-date information about this member.
		return true
	}

	// Move the faulty member over to the member list
	l.faultyMembers = slices.Delete(l.faultyMembers, faultyMemberIndex, faultyMemberIndex+1)
	l.addMember(encoding.Member{
		Address:           alive.Source,
		State:             encoding.MemberStateAlive,
		LastStateChange:   time.Now(),
		IncarnationNumber: alive.IncarnationNumber,
	})
	l.gossipQueue.Add(&alive)
	return true
}

func (l *List) handleAliveForMembers(alive gossip.MessageAlive) bool {
	memberIndex := slices.IndexFunc(l.members, func(member encoding.Member) bool {
		return member.Address.Equal(alive.Source)
	})
	if memberIndex == -1 {
		// The member is not part of our members list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if alive.IncarnationNumber <= member.IncarnationNumber {
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
	member.LastStateChange = time.Now()
	l.gossipQueue.Add(&alive)
	return true
}

func (l *List) handleAliveForUnknown(alive gossip.MessageAlive) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:           alive.Source,
		State:             encoding.MemberStateAlive,
		LastStateChange:   time.Now(),
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

	if faulty.IncarnationNumber < l.incarnationNumber {
		// We have a more up-to-date state than the gossip. Nothing to do.
		return true
	}

	// We need to re-join. Add a new alive message to gossip.
	// Also make sure that our incarnation number is bigger than before.
	l.incarnationNumber = max(l.incarnationNumber+1, faulty.IncarnationNumber+1)
	l.gossipQueue.Add(&gossip.MessageAlive{
		Source:            l.self,
		IncarnationNumber: l.incarnationNumber,
	})
	return true
}

func (l *List) handleFaultyForFaultyMembers(faulty gossip.MessageFaulty) bool {
	faultyMemberIndex := slices.IndexFunc(l.faultyMembers, func(member encoding.Member) bool {
		return member.Address.Equal(faulty.Destination)
	})
	if faultyMemberIndex == -1 {
		// The member is not part of our faulty members list. Nothing to do.
		return false
	}
	faultyMember := &l.faultyMembers[faultyMemberIndex]

	if faulty.IncarnationNumber < faultyMember.IncarnationNumber {
		// We have more up-to-date information about this member.
		return true
	}

	// Update the incarnation number to make sure we have the most current incarnation.
	faultyMember.IncarnationNumber = faulty.IncarnationNumber
	return true
}

func (l *List) handleFaultyForMembers(faulty gossip.MessageFaulty) bool {
	memberIndex := slices.IndexFunc(l.members, func(member encoding.Member) bool {
		return member.Address.Equal(faulty.Destination)
	})
	if memberIndex == -1 {
		// The member is not part of our member list. Nothing to do.
		return false
	}
	member := &l.members[memberIndex]

	if faulty.IncarnationNumber < member.IncarnationNumber {
		// We have more up-to-date information about this member.
		return true
	}

	// Remove member from member list and put it on the faulty member list.
	l.removeMemberByAddress(faulty.Destination)
	l.faultyMembers = append(l.faultyMembers, encoding.Member{
		Address:           faulty.Destination,
		State:             encoding.MemberStateFaulty,
		LastStateChange:   time.Now(),
		IncarnationNumber: faulty.IncarnationNumber,
	})
	l.gossipQueue.Add(&faulty)
	return true
}

func (l *List) handleFaultyForUnknown(faulty gossip.MessageFaulty) {
	// We don't know about this member yet. Add it to our faulty member list and gossip about it.
	l.faultyMembers = append(l.faultyMembers, encoding.Member{
		Address:           faulty.Destination,
		State:             encoding.MemberStateFaulty,
		LastStateChange:   time.Now(),
		IncarnationNumber: faulty.IncarnationNumber,
	})
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

	l.gossipQueue.PrioritizeForAddress(address)
	for i := 0; i < l.gossipQueue.Len(); i++ {
		var gossipN int
		l.datagramBuffer, gossipN, err = l.gossipQueue.Get(i).AppendToBuffer(l.datagramBuffer)
		if err != nil {
			return err
		}

		if len(l.datagramBuffer) > l.config.MaxDatagramLength {
			l.datagramBuffer = l.datagramBuffer[:len(l.datagramBuffer)-gossipN]
			break
		}

		l.gossipQueue.MarkTransmitted(i)
		datagramN += gossipN
	}

	if err := l.config.UDPClient.Send(address, l.datagramBuffer); err != nil {
		return err
	}
	return nil
}

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
	return &l.members[randomIndex]
}

func (l *List) getRandomMember() *encoding.Member {
	randomIndex := rand.Intn(len(l.members))
	return &l.members[randomIndex]
}

func (l *List) isMember(address encoding.Address) bool {
	return slices.ContainsFunc(l.members, func(member encoding.Member) bool {
		return address.Equal(member.Address)
	})
}

func (l *List) getMember(address encoding.Address) *encoding.Member {
	index := slices.IndexFunc(l.members, func(member encoding.Member) bool {
		return address.Equal(member.Address)
	})
	if index == -1 {
		return nil
	}
	return &l.members[index]
}

func (l *List) addMember(member encoding.Member) {
	l.logger.Info(
		"Member added",
		"address", member.Address,
		"incarnation-number", member.IncarnationNumber,
	)

	// We append the new member always at the end of the member list. Remember the index for later.
	memberIndex := len(l.members)
	l.members = append(l.members, member)

	// We pick a random location to insert the new member into the random indexes slice. We need to add +1 to the length
	// of that slice to allow for appending at the end.
	insertIndex := rand.Intn(len(l.randomIndexes) + 1)
	l.randomIndexes = slices.Insert(l.randomIndexes, insertIndex, memberIndex)
	if insertIndex <= l.nextRandomIndex {
		// The new member index was inserted before or at the next random index. We therefore move the next random index
		// forward by one to not have the same member be picked twice in a row.
		l.nextRandomIndex++
	}
}

func (l *List) removeMemberByAddress(address encoding.Address) {
	index := slices.IndexFunc(l.members, func(member encoding.Member) bool {
		return address.Equal(member.Address)
	})
	if index == -1 {
		return
	}
	l.removeMemberByIndex(index)
}

func (l *List) removeMemberByIndex(index int) {
	l.logger.Info(
		"Member removed",
		"address", l.members[index].Address,
		"incarnation-number", l.members[index].IncarnationNumber,
	)

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
}
