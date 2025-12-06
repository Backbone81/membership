package failuredetection

import (
	"errors"
	"fmt"
	"net"

	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
	"github.com/backbone81/membership/internal/utility"
)

var (
	minMemberCount int
	linearCutoff   int
	maxMemberCount int
)

// failureDetectionCmd represents the firstdetection command
var failureDetectionCmd = &cobra.Command{
	Use:   "failure-detection",
	Short: "How long a cluster needs to detect a failed member.",
	Long: `Simulates clusters of different sizes with one member failed.
Measures the number of protocol periods until any non-faulty member declares the failed member as faulty.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utility.CreateLogger(0)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		return Simulate(minMemberCount, linearCutoff, maxMemberCount, logger)
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(failureDetectionCmd)

	failureDetectionCmd.PersistentFlags().IntVar(
		&minMemberCount,
		"min-member-count",
		2,
		"The minimum member count to simulate.",
	)
	failureDetectionCmd.PersistentFlags().IntVar(
		&linearCutoff,
		"linear-cutoff",
		8,
		"Member counts increase linear between min-member-count and linear-cutoff. After linear-cutoff, member counts are doubled.",
	)
	failureDetectionCmd.PersistentFlags().IntVar(
		&maxMemberCount,
		"max-member-count",
		512,
		"The maximum member count to simulate.",
	)
}

// Simulate measures the time in protocol periods in which a failed member is detected by any other member.
func Simulate(minMemberCount int, linearCutoff int, maxMemberCount int, logger logr.Logger) error {
	for memberCount := range utility.ClusterSize(minMemberCount, linearCutoff, maxMemberCount) {
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
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		}
		for j := range memberCount {
			options = append(options,
				membership.WithBootstrapMember(encoding.NewAddress(net.IPv4(255, 255, 255, 255), j+1)),
			)
		}
		newList := membership.NewList(options...)
		newDebugList := membership.DebugList(newList)
		newDebugList.ClearGossip()
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
	for i := range 1024 {
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
			logger.Info("Member failure detected", "cluster-size", memberCount, "protocol-periods", i+1)
			return nil
		}
	}
	return errors.New("max number of protocol periods exceeded")
}
