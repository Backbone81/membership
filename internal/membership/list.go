package membership

import (
	"fmt"
	"math/rand"
	"slices"
	"time"
)

type List struct {
	sequenceNumber               int
	failureDetectionSubgroupSize int

	members         []Member
	randomIndexes   []int
	nextRandomIndex int
	clientTransport ClientTransport
	gossipQueue     *GossipQueue
	datagramBuilder DatagramBuilder
	self            Endpoint

	datagramBuffer      []byte
	pendingDirectProbes map[string]time.Time
}

func NewList() *List {
	return &List{}
}

func (l *List) DirectProbe() error {
	if len(l.members) < 1 {
		// As long as there are no other members, we don't have to probe anyone.
		return nil
	}

	l.sequenceNumber++
	message := MessageDirectPing{
		Source:         l.self,
		SequenceNumber: l.sequenceNumber,
	}

	var err error
	l.datagramBuffer, _, err = l.datagramBuilder.AppendToBuffer(l.datagramBuffer, &message, l.gossipQueue)
	if err != nil {
		return err
	}

	member := l.getNextMember()
	l.pendingDirectProbes[member.Endpoint.String()] = time.Now()
	if err := l.clientTransport.Send(member.Endpoint, l.datagramBuffer); err != nil {
		return err
	}

	return nil
}

func (l *List) IndirectProbe() error {
	_, err := l.pickIndirectProbes(l.failureDetectionSubgroupSize)
	if err != nil {
		return fmt.Errorf("picking indirect probe members: %w", err)
	}

	// Do an indirect ping to the member

	return nil
}

func (l *List) IndirectProbeSuccess() bool {
	return false
}

func (l *List) pickDirectProbe() (Member, error) {
	// Select a random member, which is not self.
	return Member{}, nil
}

func (l *List) pickIndirectProbes(count int) ([]Member, error) {
	// Select count random members which is not self and not member.
	return nil, nil
}

func (l *List) HandleDirectPing(message MessageDirectPing) {
}

func (l *List) HandleDirectAck(message MessageDirectAck) {
}

func (l *List) HandleIndirectPing(message MessageIndirectPing) {
}

func (l *List) HandleIndirectAck(message MessageIndirectAck) {
}

func (l *List) HandleSuspect(message MessageSuspect) {
}

func (l *List) HandleAlive(message MessageAlive) {
}

func (l *List) HandleFaulty(message MessageFaulty) {
}

func (l *List) getNextMember() *Member {
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

func (l *List) getRandomMember() *Member {
	randomIndex := rand.Intn(len(l.members))
	return &l.members[randomIndex]
}

func (l *List) isMember(endpoint Endpoint) bool {
	return slices.ContainsFunc(l.members, func(member Member) bool {
		return endpoint.Equal(member.Endpoint)
	})
}

func (l *List) getMember(endpoint Endpoint) *Member {
	index := slices.IndexFunc(l.members, func(member Member) bool {
		return endpoint.Equal(member.Endpoint)
	})
	if index == -1 {
		return nil
	}
	return &l.members[index]
}

func (l *List) addMember(member Member) {
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

func (l *List) deleteMember(endpoint Endpoint) {
	index := slices.IndexFunc(l.members, func(member Member) bool {
		return endpoint.Equal(member.Endpoint)
	})
	if index == -1 {
		return
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
}
