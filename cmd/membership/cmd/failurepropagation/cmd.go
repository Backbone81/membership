package failurepropagation

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"slices"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/backbone81/membership/internal/transport"
	"github.com/backbone81/membership/internal/utility"
)

var (
	minMemberCount int
	linearCutoff   int
	maxMemberCount int
)

// failurePropagationCmd represents the allDetection command.
var failurePropagationCmd = &cobra.Command{
	Use:   "failure-propagation",
	Short: "How long a cluster needs to propagate a failed member.",
	Long: `Simulates clusters of different sizes with one member failed.
Measures the number of protocol periods until all non-faulty members know about the failed member.
Note that this simulation does not execute the periodic full list sync which the default membership list would do.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

		return Simulate(minMemberCount, linearCutoff, maxMemberCount, logger)
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(failurePropagationCmd)

	failurePropagationCmd.PersistentFlags().IntVar(
		&minMemberCount,
		"min-member-count",
		2,
		"The minimum member count to simulate.",
	)
	failurePropagationCmd.PersistentFlags().IntVar(
		&linearCutoff,
		"linear-cutoff",
		8,
		"Member counts increase linear between min-member-count and linear-cutoff. After linear-cutoff, member counts are doubled.",
	)
	failurePropagationCmd.PersistentFlags().IntVar(
		&maxMemberCount,
		"max-member-count",
		512,
		"The maximum member count to simulate.",
	)
}

// Simulate measures the time in protocol periods in which a failed member is known to all other members.
// Either by ping other by gossip.
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
	lists := make([]*membership.List, 0, memberCount)
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

//nolint:gocognit,cyclop
func runProtocol(logger logr.Logger, lists []*membership.List, memoryTransport *transport.Memory, memberCount int) error {
	detected := make([]int, len(lists))
	for i := range detected {
		detected[i] = math.MaxInt
	}
	var detectedCount int
	for protocolPeriod := range 1024 {
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
		for listIndex, list := range lists {
			if list.Len() == memberCount-1 {
				continue
			}
			if protocolPeriod < detected[listIndex] {
				detected[listIndex] = protocolPeriod
				detectedCount++
			}
		}
		if detectedCount == len(detected) {
			break
		}
	}
	if detectedCount != len(detected) {
		return errors.New("max number of protocol periods exceeded")
	}
	slices.Sort(detected)
	logger.Info(
		"Member failure propagated",
		"cluster-size", memberCount,
		"min", detected[0]+1,
		"median", detected[len(detected)/2]+1,
		"max", detected[len(detected)-1]+1,
	)
	return nil
}
