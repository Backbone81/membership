package firstdetection

import (
	"fmt"
	"net"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
	"github.com/backbone81/membership/internal/utility"
	"github.com/go-logr/logr"
)

// FirstDetection measures the protocol period in which one of the members detects the failure of another member.
// This is what the SWIM paper provides in section "5.2. Detection and Dissemination Latency of Membership Updates" in
// "Figure 3(a)".
// The expectation is that the first detection should be around 1.58 with some variation around that based on
// randomness.
func FirstDetection(logger logr.Logger) error {
	logger.Info("The number of protocol periods between a member failure and its first detection at some non-faulty member.")
	for memberCount := range utility.ClusterSize(2, 64, 512) {
		memoryTransport := transport.NewMemory()

		lists, err := buildCluster(memberCount, memoryTransport)
		if err != nil {
			return err
		}
		if err := runProtocol(logger, lists, memoryTransport, memberCount); err != nil {
			return err
		}
	}
	return nil
}

func buildCluster(memberCount int, memoryTransport *transport.Memory) ([]*membership.List, error) {
	// Create our membership lists and make them know each other. Note that we are adding one member less but
	// still force all members to know of the last one which we do not add. This is the member we simulate to be
	// faulty.
	var lists []*membership.List
	for i := range memberCount - 1 {
		address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), i+1)
		options := []membership.Option{
			membership.WithLogger(logr.Discard()),
			membership.WithAdvertisedAddress(address),
			membership.WithUDPClient(memoryTransport.Client()),
			membership.WithTCPClient(memoryTransport.Client()),
			membership.WithIndirectPingMemberCount(1),
			// We set the safety factor to 0 which results in the protocol immediately declaring a member as failed
			// as soon as a ping fails during a protocol period.
			membership.WithSafetyFactor(0),
		}
		for j := range memberCount {
			options = append(options,
				membership.WithBootstrapMember(encoding.NewAddress(net.IPv4(255, 255, 255, 255), j+1)),
			)
		}
		newList := membership.NewList(options...)
		newList.ClearGossip()
		memoryTransport.AddTarget(address, newList)
		lists = append(lists, newList)
	}

	// Make sure that all membership lists have the correct number of members.
	for i, list := range lists {
		if list.Len() != memberCount-1 {
			return nil, fmt.Errorf("expected member list %d to have %d members but got %d", i, memberCount-1, list.Len())
		}
	}
	return lists, nil
}

func runProtocol(logger logr.Logger, lists []*membership.List, memoryTransport *transport.Memory, memberCount int) error {
	for i := range 16 {
		for _, list := range lists {
			if err := list.DirectPing(); err != nil {
				return err
			}
		}
		if err := memoryTransport.FlushAllPendingSends(); err != nil {
			return err
		}

		for _, list := range lists {
			if err := list.IndirectPing(); err != nil {
				return err
			}
		}
		if err := memoryTransport.FlushAllPendingSends(); err != nil {
			return err
		}

		for _, list := range lists {
			if err := list.EndOfProtocolPeriod(); err != nil {
				return err
			}
		}
		for _, list := range lists {
			if list.Len() == memberCount-1 {
				continue
			}
			logger.Info("Member failure detected", "group-size", memberCount, "protocol-periods", i+1)
			return nil
		}
	}
	return fmt.Errorf("max number of protocol periods exceeded")
}
