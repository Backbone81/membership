package cmd

import (
	"fmt"
	"math"
	"net"
	"os"

	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
	"github.com/backbone81/membership/internal/utility"
)

var (
	verbosity                int
	maxDatagramLengthSend    int
	maxDatagramLengthReceive int
	memberCount              int
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "single-join",
	Short:        "Lets a single new member join an established cluster.",
	Long:         `Lets a single new member join an established cluster.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup the logger
		logger, zapLogger, err := utility.CreateLogger(verbosity)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		// Create the cluster with all initial members.
		memoryTransport := transport.NewMemory()
		var lists []*membership.List
		for i := range memberCount {
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), i+1)
			options := []membership.Option{
				membership.WithLogger(logger.WithValues("list", address)),
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
			membership.DebugList(newList).ClearGossip()
			memoryTransport.AddTarget(address, newList)
			lists = append(lists, newList)
		}

		// Add the member which is now joining the existing cluster.
		address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), math.MaxUint16)
		options := []membership.Option{
			membership.WithLogger(logger.WithValues("list", address)),
			membership.WithAdvertisedAddress(address),
			membership.WithUDPClient(memoryTransport.Client()),
			membership.WithTCPClient(memoryTransport.Client()),
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		}
		options = append(options,
			membership.WithBootstrapMember(encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)),
		)
		newList := membership.NewList(options...)
		memoryTransport.AddTarget(address, newList)
		lists = append(lists, newList)

		periodCount := int(math.Ceil(utility.DisseminationPeriods(membership.DefaultConfig.SafetyFactor, len(lists))))
		for i := range periodCount {
			logger.Info("> Start of protocol period", "period", i)
			if err := dumpClusterState(lists, i); err != nil {
				return err
			}

			logger.Info("> Executing direct pings")
			for _, list := range lists {
				if err := list.DirectPing(); err != nil {
					return err
				}
			}
			if err := memoryTransport.FlushAllPendingSends(); err != nil {
				return err
			}

			logger.Info("> Executing indirect pings")
			for _, list := range lists {
				if err := list.IndirectPing(); err != nil {
					return err
				}
			}
			if err := memoryTransport.FlushAllPendingSends(); err != nil {
				return err
			}

			logger.Info("> Executing end of protocol period")
			for _, list := range lists {
				if err := list.EndOfProtocolPeriod(); err != nil {
					return err
				}
			}
		}

		for i, list := range lists[:len(lists)-1] {
			if len(list.Get()) != memberCount {
				return fmt.Errorf("expected list %d to have %d members but got %d", i+1, memberCount, len(list.Get()))
			}
		}
		return nil
	},
}

func dumpClusterState(lists []*membership.List, protocolPeriod int) error {
	file, err := os.Create(fmt.Sprintf("membership-period-%d.txt", protocolPeriod))
	if err != nil {
		return err
	}
	defer file.Close()

	for _, list := range lists {
		if err := membership.DebugList(list).WriteInternalDebugState(file); err != nil {
			return err
		}
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(
		&verbosity,
		"verbosity",
		"v",
		0,
		"Sets the verbosity for log output. 0 reports info and error messages, while 1 and up report more detailed logs.",
	)
	rootCmd.PersistentFlags().IntVar(
		&maxDatagramLengthSend,
		"max-datagram-length-send",
		512,
		`The maximum length of network messages in bytes. This should be set to a value which does not cause fragmentation.
This value must be equal or smaller to max-datagram-length-receive to not cause data loss.
A conservative length with most compatibility is (576 bytes IP datagram length) - (20 to 60 bytes IP header) - (8 bytes UDP header).
A progressive length for an internal ethernet based network is (1500 bytes ethernet MTU) - (20 to 60 bytes IP header) - (8 bytes UDP header).`,
	)
	rootCmd.PersistentFlags().IntVar(
		&maxDatagramLengthReceive,
		"max-datagram-length-receive",
		512,
		`The maximum length of network messages in bytes. This should be set to a value which does not cause fragmentation.
The value must be equal or bigger to max-datagram-length-send to not cause data loss.
A conservative length with most compatibility is (576 bytes IP datagram length) - (20 to 60 bytes IP header) - (8 bytes UDP header).
A progressive length for an internal ethernet based network is (1500 bytes ethernet MTU) - (20 to 60 bytes IP header) - (8 bytes UDP header).`,
	)

	rootCmd.PersistentFlags().IntVar(
		&memberCount,
		"member-count",
		64,
		`The number of members the initial cluster should consist of.`,
	)
}
