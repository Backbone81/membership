package lossyjoin

import (
	"math"
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
	memberCount        int
	networkReliability float64
)

// lossyJoinCmd represents the allDetection command
var lossyJoinCmd = &cobra.Command{
	Use:   "lossy-join",
	Short: "Joins a set of new members through a lossy network.",
	Long: `Starts off with a single member and joins one new member every protocol period. The network is unreliable
and drops some network messages. The simulation shows at what number of members the cluster stabilizes.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utility.CreateLogger(0)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		return runProtocol(logger, transport.NewMemory(), memberCount)
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(lossyJoinCmd)

	lossyJoinCmd.PersistentFlags().IntVar(
		&memberCount,
		"member-count",
		512,
		"The member count to simulate.",
	)
	lossyJoinCmd.PersistentFlags().Float64Var(
		&networkReliability,
		"network-reliability",
		0.99,
		"The probability of a network message arriving at its target as a value between 0 and 1.",
	)
}

func runProtocol(logger logr.Logger, memoryTransport *transport.Memory, memberCount int) error {
	var observedMinMemberCount []int
	var lists []*membership.List
	for protocolPeriod := range 1024 {
		if protocolPeriod < memberCount {
			// We join one member in every protocol period until the max member count is reached.
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), protocolPeriod+1)
			options := []membership.Option{
				membership.WithLogger(logr.Discard()),
				membership.WithAdvertisedAddress(address),
				membership.WithUDPClient(&transport.Unreliable{
					Transport:   memoryTransport.Client(),
					Reliability: networkReliability,
				}),
				membership.WithTCPClient(&transport.Unreliable{
					Transport:   memoryTransport.Client(),
					Reliability: networkReliability,
				}),
				membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
			}
			for _, list := range lists {
				options = append(options,
					membership.WithBootstrapMember(list.Config().AdvertisedAddress),
				)
			}
			newList := membership.NewList(options...)
			memoryTransport.AddTarget(address, newList)
			lists = append(lists, newList)
		}

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

		minMemberCount := math.MaxInt
		for _, list := range lists {
			minMemberCount = min(minMemberCount, list.Len())
		}
		observedMinMemberCount = append(observedMinMemberCount, minMemberCount)

		logger.Info(
			"Cluster members",
			"cluster-size", memberCount,
			"protocol-period", protocolPeriod+1,
			"min-cluster-size", minMemberCount+1,
		)
	}
	return nil
}
