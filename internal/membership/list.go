package membership

import (
	"errors"
	"fmt"
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

	pickRandomMembersResult []*encoding.Member
	pickRandomMembersSwap   map[int]int
}

// NewList creates a new membership list.
func NewList(options ...Option) *List {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}

	if config.UDPClient == nil {
		panic("you must provide a UDP client")
	}
	if config.TCPClient == nil {
		panic("you must provide a TCP client")
	}
	if config.RoundTripTimeTracker == nil {
		panic("you must provide a round trip time tracker")
	}

	newList := List{
		config:                  config,
		logger:                  config.Logger,
		self:                    config.AdvertisedAddress,
		gossipQueue:             gossip.NewQueue(),
		datagramBuffer:          make([]byte, 0, config.MaxDatagramLengthSend),
		pendingDirectPings:      make([]PendingDirectPing, 0, 16),
		pendingDirectPingsNext:  make([]PendingDirectPing, 0, 16),
		pendingIndirectPings:    make([]PendingIndirectPing, 0, 16),
		pickRandomMembersResult: make([]*encoding.Member, 0, 16),
		pickRandomMembersSwap:   make(map[int]int, 16),
	}

	// We need to gossip our own alive. Otherwise, nobody will pick us up into their own member list.
	newList.gossipQueue.Add(encoding.MessageAlive{
		Destination:       config.AdvertisedAddress,
		IncarnationNumber: 0,
	}.ToMessage())
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

// ForEach executes the given function for all address of all members stored in the list. The members are sorted by
// address ascending.
// While iterating over all members, the internal mutex is locked. Do not execute lengthy operations while iterating
// over all members, as that will block processing of network messages. In addition, do not call any other method
// of List during iteration, as that will cause a deadlock. Create your own copy of the member list and work on that
// list if you need to execute long-running operations or need to call methods on List.
//
// Note that we are explicitly not providing a range over function type for iterating over all member addresses, because
// that would cause memory allocations for the range over for loop, as it needs to introduce state which is allocated
// on the heap. The solution with ForEach is less nice, but it allows for zero allocations.
func (l *List) ForEach(fn func(encoding.Address) bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, member := range l.members {
		if !fn(member.Address) {
			return
		}
	}
}

// DirectPing executes the first step in the SWIM protocol by directly pinging other members.
func (l *List) DirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// As we are supporting to directly ping multiple members, we need to make sure that we are not exceeding the
	// current member count. Otherwise, we would ping the same member multiple times in the same protocol period.
	for range min(len(l.members), l.config.DirectPingMemberCount) {
		directPing := encoding.MessageDirectPing{
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
		if err := l.sendWithGossip(destination, directPing.ToMessage()); err != nil {
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

// sendWithGossip sends the given network message to the given address, It fills up the remaining space in the datagram
// with gossip from the gossip queue.
func (l *List) sendWithGossip(address encoding.Address, message encoding.Message) error {
	l.datagramBuffer = l.datagramBuffer[:0]

	var err error
	var datagramN int
	l.datagramBuffer, datagramN, err = message.AppendToBuffer(l.datagramBuffer)
	if err != nil {
		return err
	}

	// Make sure that we send gossip about our destination first, to allow quicker refutation of suspects.
	l.gossipQueue.Prioritize(address)

	var gossipAdded int
	for _, msg := range l.gossipQueue.All() {
		var gossipN int
		l.datagramBuffer, gossipN, err = msg.AppendToBuffer(l.datagramBuffer)
		if err != nil {
			return err
		}

		if len(l.datagramBuffer) > l.config.MaxDatagramLengthSend {
			// Appending the last gossip exceeded the maximum size we want to have for our network message. Reset the
			// buffer back to its size before we added the last gossip.
			l.datagramBuffer = l.datagramBuffer[:datagramN]
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

// IndirectPing executes the second step in the SWIM protocol by requesting indirect pings for direct pings which timed
// out.
func (l *List) IndirectPing() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// An indirect ping only makes sense whe we have at least two members.
	if len(l.members) < 2 {
		return nil
	}

	// As we want to do the indirect pings for all pending direct pings, we collect all errors and return them as
	// a joined error at the end. Otherwise, the first error would stop the iteration and break the expected behavior.
	var joinedErr error
	for _, directPing := range l.pendingDirectPings {
		if !directPing.MessageIndirectPing.IsZero() {
			// We are not interested in direct pings which we do as a request for an indirect ping. We only do
			// indirect pings for direct pings we initiated on our own.
			continue
		}

		indirectPing := encoding.MessageIndirectPing{
			Source:      l.self,
			Destination: directPing.Destination,

			// We use the same sequence number as we did for the corresponding direct ping. That way, logs can
			// correlate the indirect ping with the direct ping, and we know which indirect pings to discard when
			// the direct ping succeeds late.
			SequenceNumber: directPing.MessageDirectPing.SequenceNumber,
		}

		// Send the indirect pings to the indirect ping members and join up all errors which might occur.
		members := l.pickRandomMembersWithout(l.config.IndirectPingMemberCount, directPing.Destination)
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
			if err := l.sendWithGossip(member.Address, indirectPing.ToMessage()); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		}
	}
	return joinedErr
}

// pickRandomMembers returns count random members using a partial Fisher-Yates shuffle
// (https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle). This avoids copying and shuffling the entire members
// slice and is basically the same what rand.Shuffle is using for shuffling full slices.
// As we try to avoid copying the full slice, we use a map to remember those swaps we already did. This allows us to
// have minimal memory consumption even for large member lists.
// WARNING: The member slice returned by this method is only valid until the next call to pickRandomMembers or
// pickRandomMembersWithout. Create a copy if you need to retain the result longer.
// TODO: We might think about replacing the returned slice with a range over function. This could free us from keeping
// the return slice around to avoid memory allocations. pickRandomMembersWithout can the also be implemented as a filter
// over that range over function. The swap map still needs to remain, but everything else would be simpler.
func (l *List) pickRandomMembers(count int) []*encoding.Member {
	count = min(count, len(l.members))
	if count == 0 {
		return nil
	}

	// Reset the result slice and reset the map.
	l.pickRandomMembersResult = l.pickRandomMembersResult[:0]
	clear(l.pickRandomMembersSwap)

	// We iterate over the number of elements we want to retrieve.
	for i := 0; i < count; i++ {
		// For every element, we pick a random other element which is identical to the current element or bigger.
		j := i + rand.Intn(len(l.members)-i)

		// We look up the real indexes according to what swaps we already did in the past.
		iReal := l.pickRandomMemberIndex(i)
		jReal := l.pickRandomMemberIndex(j)

		// Let's remember that the j element is now replaced by the real i element. Note that we do not remember the
		// i element, because we will never look at it again, we are only swapping with elements to the right of i, not
		// left of i.
		l.pickRandomMembersSwap[j] = iReal

		// Append the swapped member to the result.
		l.pickRandomMembersResult = append(l.pickRandomMembersResult, &l.members[jReal])
	}
	return l.pickRandomMembersResult
}

// pickRandomMemberIndex is a helper method which resolves a given index through the swap map to get the real index.
func (l *List) pickRandomMemberIndex(index int) int {
	if value, found := l.pickRandomMembersSwap[index]; found {
		return value
	}
	return index
}

// pickRandomMembersWithout returns count random members, excluding the member with the given address.
// If the excluded address is selected, it's replaced with an additional random member.
// WARNING: The member slice returned by this method is only valid until the next call to pickRandomMembers or
// pickRandomMembersWithout. Create a copy if you need to retain the result longer.
func (l *List) pickRandomMembersWithout(count int, exclude encoding.Address) []*encoding.Member {
	// We retrieve one member more that requested to allow for an additional member if we find the excluded address.
	result := l.pickRandomMembers(count + 1)

	// Find and remove the excluded address if present
	for i, member := range result {
		if member.Address.Equal(exclude) {
			return utility.SwapDelete(result, i)
		}
	}

	// Ensure we don't return more than requested.
	if len(result) > count {
		result = result[:count]
	}
	return result
}

// EndOfProtocolPeriod is the last step in the SWIM protocol where we check which pings went unanswered, and we declare
// as suspect or faulty which needs declaring.
func (l *List) EndOfProtocolPeriod() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Adjust the gossip queue to the potentially changed cluster size. Either keep gossip longer because of bigger
	// cluster or keep gossip shorter, because of smaller cluster.
	l.gossipQueue.SetMaxTransmissionCount(l.requiredDisseminationPeriods())

	// We first process failed pings which lead to suspect declarations, and then mark suspects as faulty. This allows
	// us to declare suspect and faulty within the same protocol period, if needed. This is helpful for tests and
	// benchmarks where we can only observe the list state through the public interface.
	l.processFailedPings()
	l.markSuspectsAsFaulty()

	// TODO: We do not account for suspect members right now, because that would require scanning the whole member
	// list. We extend the suspect metric as soon as we have dedicated suspect member tracking (we have some other todo)
	// for that already.
	MembersByState.WithLabelValues("alive").Set(float64(len(l.members)))
	MembersByState.WithLabelValues("faulty").Set(float64(len(l.faultyMembers)))
	return nil
}

// requiredDisseminationPeriods returns the number of protocol periods which are deemed safe for disseminating gossip
// messages and for declaring as suspect as faulty. Note that a safety factor of 0 will always lead to 0 periods
// causing instant suspect and faulty declarations.
func (l *List) requiredDisseminationPeriods() int {
	return int(math.Ceil(utility.DisseminationPeriods(l.config.SafetyFactor, len(l.members))))
}

// processFailedPings loops through all pending direct pings, marks members as suspect which did not answer to pings and
// adds a gossip message about that suspect message.
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
		MemberStateTransitionsTotal.WithLabelValues("declared_suspect").Inc()

		// We need to mark the member as suspect and gossip about it.
		member.State = encoding.MemberStateSuspect
		member.SuspicionPeriodCounter = 0
		l.gossipQueue.Add(encoding.MessageSuspect{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		}.ToMessage())
	}

	// We swap the pending direct pings of the current protocol period with the pending direct pings of the next
	// protocol period - resetting the pings for the next period to nothing. This allows us to re-use the same slices
	// over and over without new allocations.
	l.pendingDirectPings, l.pendingDirectPingsNext = l.pendingDirectPingsNext, l.pendingDirectPings[:0]

	// As indirect pings always happen with a direct ping not being satisfied before, we can clear the indirect pings
	// without any further actions, as those actions have already been taken on the pending direct pings.
	l.pendingIndirectPings = l.pendingIndirectPings[:0]
}

// markSuspectsAsFaulty loops through all members and increases the suspect counter on each suspect. It declares members
// as faulty if they exceeded the suspicion threshold.
func (l *List) markSuspectsAsFaulty() {
	suspicionPeriodThreshold := l.requiredDisseminationPeriods()

	// As we are potentially removing elements from the member list, we need to iterate from the back to the front in
	// order to not skip a member when the content changes.
	// TODO: Iterating over all members to search for suspects is wasteful. Introduce a suspectCounters map and remove
	// the field SuspicionPeriodCounter from the member struct. You need to update the indexes in this map when members
	// are added or removed, and when members are declared suspect or alive.
	for i := len(l.members) - 1; i >= 0; i-- {
		member := &l.members[i]
		if member.State != encoding.MemberStateSuspect {
			// We are only interested in looking at suspect members.
			continue
		}
		member.SuspicionPeriodCounter++
		if member.SuspicionPeriodCounter <= suspicionPeriodThreshold {
			// We are only interested in members exceeding the suspicion threshold.
			continue
		}

		l.logger.Info(
			"Member declared as faulty",
			"source", l.self,
			"destination", member.Address,
			"incarnation-number", member.IncarnationNumber,
		)
		MemberStateTransitionsTotal.WithLabelValues("declared_faulty").Inc()

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
		l.gossipQueue.Add(encoding.MessageFaulty{
			Source:            l.self,
			Destination:       member.Address,
			IncarnationNumber: member.IncarnationNumber,
		}.ToMessage())
		l.removeMemberByIndex(i) // must always happen last to keep the member alive during this method
	}
}

// RequestList is a protocol step which is not part of the core SWIM protocol, but it is required for letting new
// members know the full member list quickly, and it is helpful in addressing some randomness issues which might lead
// to some member not getting the gossip about a specific change. RequestList picks one random member and requests
// a full list of all alive, suspect and faulty members. The request is sent as a standard datagram with gossip, while
// the response is returned as TCP message. This operation can be expensive in time and space and should be executed
// at a much lower frequency compared to the standard SWIM actions.
func (l *List) RequestList() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	members := l.pickRandomMembers(1)
	if len(members) < 1 {
		// We could not pick a member to request the member list from. There probably is no other member at all.
		return nil
	}

	listRequest := encoding.MessageListRequest{
		Source: l.self,
	}

	l.logger.V(1).Info(
		"Requesting member list",
		"destination", members[0].Address,
	)
	if err := l.sendWithGossip(members[0].Address, listRequest.ToMessage()); err != nil {
		return err
	}
	return nil
}

// BroadcastShutdown is picking some members at random and sends those a faulty message about itself. This helps in
// disseminating graceful shutdowns a lot quicker than waiting for a ping to fail and then to wait through a suspect
// timeout.
func (l *List) BroadcastShutdown() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	faultyMessage := encoding.MessageFaulty{
		Source:            l.self,
		Destination:       l.self,
		IncarnationNumber: l.incarnationNumber,
	}

	members := l.pickRandomMembers(l.config.ShutdownMemberCount)
	for _, member := range members {
		l.logger.Info(
			"Broadcasting shutdown",
			"address", member.Address,
		)
		// We send our broadcast with the gossip we have, to help disseminate that information before we are gone.
		if err := l.sendWithGossip(member.Address, faultyMessage.ToMessage()); err != nil {
			return err
		}
	}
	return nil
}

// addMember adds the given member as a new member. It updates all bookkeeping which might be affected by this change.
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
	MemberStateTransitionsTotal.WithLabelValues("added").Inc()
}

// removeMemberByAddress removes the member with the given address from the list of members. Updating the relevant
// bookkeeping at the same time.
func (l *List) removeMemberByAddress(address encoding.Address) {
	index, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: address},
		encoding.CompareMember,
	)
	if !found {
		return
	}
	l.removeMemberByIndex(index)
}

// removeMemberByIndex removes the member with the given index from the list of members. Updating the relevant
// bookkeeping at the same time.
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
	MemberStateTransitionsTotal.WithLabelValues("removed").Inc()
}

// DispatchDatagram is the entrypoint which processes messages received by other members. The buffer provided as
// parameter might contain any number of messages. This method will unmarshal messages and pass them on for processing
// until the buffer is exhausted.
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
			MessagesReceivedTotal.WithLabelValues("direct_ping").Inc()
			var message encoding.MessageDirectPing
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleDirectPing(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeDirectAck:
			MessagesReceivedTotal.WithLabelValues("direct_ack").Inc()
			var message encoding.MessageDirectAck
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleDirectAck(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeIndirectPing:
			MessagesReceivedTotal.WithLabelValues("indirect_ping").Inc()
			var message encoding.MessageIndirectPing
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleIndirectPing(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeIndirectAck:
			MessagesReceivedTotal.WithLabelValues("indirect_ack").Inc()
			var message encoding.MessageIndirectAck
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleIndirectAck(message)
		case encoding.MessageTypeSuspect:
			MessagesReceivedTotal.WithLabelValues("suspect").Inc()
			var message encoding.MessageSuspect
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleSuspect(message)
		case encoding.MessageTypeAlive:
			MessagesReceivedTotal.WithLabelValues("alive").Inc()
			var message encoding.MessageAlive
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleAlive(message)
		case encoding.MessageTypeFaulty:
			MessagesReceivedTotal.WithLabelValues("faulty").Inc()
			var message encoding.MessageFaulty
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			l.handleFaulty(message)
		case encoding.MessageTypeListRequest:
			MessagesReceivedTotal.WithLabelValues("list_request").Inc()
			var message encoding.MessageListRequest
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleListRequest(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		case encoding.MessageTypeListResponse:
			MessagesReceivedTotal.WithLabelValues("list_response").Inc()
			var message encoding.MessageListResponse
			n, err := message.FromBuffer(buffer)
			if err != nil {
				return err
			}
			buffer = buffer[n:]
			if err := l.handleListResponse(message); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		default:
			l.logger.Error(
				fmt.Errorf("unknown message type %d", messageType),
				"The network message has an unknown message type.",
			)
		}
	}
	return joinedErr
}

func (l *List) handleDirectPing(directPing encoding.MessageDirectPing) error {
	l.logger.V(2).Info(
		"Received direct ping",
		"source", directPing.Source,
		"sequence-number", directPing.SequenceNumber,
	)
	directAck := encoding.MessageDirectAck{
		Source:         l.self,
		SequenceNumber: directPing.SequenceNumber,
	}
	if err := l.sendWithGossip(directPing.Source, directAck.ToMessage()); err != nil {
		return err
	}
	return nil
}

func (l *List) handleDirectAck(directAck encoding.MessageDirectAck) error {
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

func (l *List) handleDirectAckForPendingDirectPings(pendingDirectPings []PendingDirectPing, directAck encoding.MessageDirectAck) ([]PendingDirectPing, error) {
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

	indirectAck := encoding.MessageIndirectAck{
		Source:         directAck.Source,
		SequenceNumber: pendingDirectPing.MessageIndirectPing.SequenceNumber,
	}
	if err := l.sendWithGossip(pendingDirectPing.MessageIndirectPing.Source, indirectAck.ToMessage()); err != nil {
		return pendingDirectPings, err
	}
	return pendingDirectPings, nil
}

func (l *List) handleDirectAckForPendingIndirectPings(directAck encoding.MessageDirectAck) {
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

func (l *List) handleIndirectPing(indirectPing encoding.MessageIndirectPing) error {
	l.logger.V(2).Info(
		"Received indirect ping",
		"source", indirectPing.Source,
		"destination", indirectPing.Destination,
		"sequence-number", indirectPing.SequenceNumber,
	)
	directPing := encoding.MessageDirectPing{
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

	if err := l.sendWithGossip(indirectPing.Destination, directPing.ToMessage()); err != nil {
		return err
	}
	return nil
}

func (l *List) handleIndirectAck(indirectAck encoding.MessageIndirectAck) {
	l.logger.V(2).Info(
		"Received indirect ack",
		"source", indirectAck.Source,
		"sequence-number", indirectAck.SequenceNumber,
	)
	l.handleIndirectAckForPendingDirectPings(indirectAck)
	l.handleIndirectAckForPendingIndirectPings(indirectAck)
}

func (l *List) handleIndirectAckForPendingDirectPings(indirectAck encoding.MessageIndirectAck) {
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

func (l *List) handleIndirectAckForPendingIndirectPings(indirectAck encoding.MessageIndirectAck) {
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

func (l *List) handleSuspect(suspect encoding.MessageSuspect) {
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

func (l *List) handleSuspectForSelf(suspect encoding.MessageSuspect) bool {
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
	l.gossipQueue.Add(encoding.MessageAlive{
		Destination:       l.self,
		IncarnationNumber: l.incarnationNumber,
	}.ToMessage())

	l.logger.Info(
		"Refuted gossip about being suspect",
		"incarnation-number", l.incarnationNumber,
	)
	MemberStateTransitionsTotal.WithLabelValues("refuted_suspect").Inc()
	return true
}

func (l *List) handleSuspectForFaultyMembers(suspect encoding.MessageSuspect) bool {
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
	l.gossipQueue.Add(suspect.ToMessage())
	return true
}

func (l *List) handleSuspectForMembers(suspect encoding.MessageSuspect) bool {
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
	l.gossipQueue.Add(suspect.ToMessage())
	return true
}

func (l *List) handleSuspectForUnknown(suspect encoding.MessageSuspect) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:                suspect.Destination,
		State:                  encoding.MemberStateSuspect,
		SuspicionPeriodCounter: 0,
		IncarnationNumber:      suspect.IncarnationNumber,
	})
	l.gossipQueue.Add(suspect.ToMessage())
}

func (l *List) handleAlive(alive encoding.MessageAlive) {
	l.logger.V(3).Info(
		"Received gossip about alive",
		"source", alive.Destination,
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

func (l *List) handleAliveForSelf(alive encoding.MessageAlive) bool {
	if !alive.Destination.Equal(l.self) {
		return false
	}

	if utility.IncarnationLessThan(alive.IncarnationNumber, l.incarnationNumber) {
		// We have a more up-to-date state than the gossip. Nothing to do.
		return true
	}

	// We need to update the incarnation number about ourselves. Add a new alive message to gossip.
	// Also make sure that our incarnation number is bigger than before.
	l.incarnationNumber = utility.IncarnationMax(l.incarnationNumber+1, alive.IncarnationNumber+1)
	l.gossipQueue.Add(encoding.MessageAlive{
		Destination:       l.self,
		IncarnationNumber: l.incarnationNumber,
	}.ToMessage())

	l.logger.Info(
		"Refuted gossip about being alive",
		"incarnation-number", l.incarnationNumber,
	)
	MemberStateTransitionsTotal.WithLabelValues("refuted_alive").Inc()
	return true
}

func (l *List) handleAliveForFaultyMembers(alive encoding.MessageAlive) bool {
	faultyMemberIndex, found := slices.BinarySearchFunc(
		l.faultyMembers,
		encoding.Member{Address: alive.Destination},
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
		Address:           alive.Destination,
		State:             encoding.MemberStateAlive,
		IncarnationNumber: alive.IncarnationNumber,
	})
	l.gossipQueue.Add(alive.ToMessage())
	return true
}

func (l *List) handleAliveForMembers(alive encoding.MessageAlive) bool {
	memberIndex, found := slices.BinarySearchFunc(
		l.members,
		encoding.Member{Address: alive.Destination},
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
	l.gossipQueue.Add(alive.ToMessage())
	return true
}

func (l *List) handleAliveForUnknown(alive encoding.MessageAlive) {
	// We don't know about this member yet. Add it to our member list and gossip about it.
	l.addMember(encoding.Member{
		Address:           alive.Destination,
		State:             encoding.MemberStateAlive,
		IncarnationNumber: alive.IncarnationNumber,
	})
	l.gossipQueue.Add(alive.ToMessage())
}

func (l *List) handleFaulty(faulty encoding.MessageFaulty) {
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

func (l *List) handleFaultyForSelf(faulty encoding.MessageFaulty) bool {
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
	l.gossipQueue.Add(encoding.MessageAlive{
		Destination:       l.self,
		IncarnationNumber: l.incarnationNumber,
	}.ToMessage())

	l.logger.Info(
		"Refuted gossip about being faulty",
		"incarnation-number", l.incarnationNumber,
	)
	MemberStateTransitionsTotal.WithLabelValues("refuted_faulty").Inc()
	return true
}

func (l *List) handleFaultyForFaultyMembers(faulty encoding.MessageFaulty) bool {
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

func (l *List) handleFaultyForMembers(faulty encoding.MessageFaulty) bool {
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
	l.gossipQueue.Add(faulty.ToMessage())
	l.removeMemberByAddress(faulty.Destination) // must always happen last to keep the member alive during this method
	return true
}

func (l *List) handleFaultyForUnknown(faulty encoding.MessageFaulty) {
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
	l.gossipQueue.Add(faulty.ToMessage())
}

func (l *List) handleListRequest(listRequest encoding.MessageListRequest) error {
	l.logger.V(2).Info(
		"Received list request",
		"source", listRequest.Source,
	)

	listResponse := encoding.MessageListResponse{
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

func (l *List) handleListResponse(listResponse encoding.MessageListResponse) error {
	l.logger.V(2).Info(
		"Received list response",
		"source", listResponse.Source,
	)
	for _, member := range listResponse.Members {
		switch member.State {
		case encoding.MemberStateAlive:
			l.handleAlive(encoding.MessageAlive{
				Destination:       member.Address,
				IncarnationNumber: member.IncarnationNumber,
			})
		case encoding.MemberStateSuspect:
			l.handleSuspect(encoding.MessageSuspect{
				Source:            listResponse.Source,
				Destination:       member.Address,
				IncarnationNumber: member.IncarnationNumber,
			})
		case encoding.MemberStateFaulty:
			l.handleFaulty(encoding.MessageFaulty{
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
