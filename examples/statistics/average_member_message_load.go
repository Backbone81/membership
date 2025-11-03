package main

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
	"github.com/go-logr/logr"
)

// AverageMemberMessageLoad calculates the average member message load per protocol period with average sends and
// receives with standard deviation. This is what the SWIM paper provides in section "5.1. Message Loads" in "Figure 2".
// The expectation is that the average should be around 2.0 with a standard deviation of about 1.0.
// TODO: The implementation is incomplete. We need to actually track the sends and receives as well as calculate the
// mean and standard deviation.
func AverageMemberMessageLoad(logger logr.Logger, maxMemberCount int) error {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	defer writer.Flush()

	if _, err := fmt.Fprintln(writer, "Group Size\tMin Sent Msgs\tAvg Sent Msgs\tMax Sent Msgs\tMin Rcv Msgs\tAvg Recv Msgs\tMax Rcv Msgs"); err != nil {
		return err
	}
	for memberCount := 2; memberCount <= maxMemberCount; memberCount++ {
		memoryTransport := transport.NewMemory()

		// Create our membership lists and make them know each other.
		var lists []*membership.List
		for i := range memberCount {
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), i+1)
			options := []membership.Option{
				membership.WithLogger(logger.WithValues("list", address)),
				membership.WithAdvertisedAddress(address),
				membership.WithUDPClient(memoryTransport.Client()),
				membership.WithTCPClient(memoryTransport.Client()),
				membership.WithIndirectPingMemberCount(1),
				membership.WithSafetyFactor(3),
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
				return fmt.Errorf("expected member list %d to have %d members but got %d", i, memberCount-1, list.Len())
			}
		}

		// Run protocol periods
		for range 40 {
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
		}
	}
	return nil
}
