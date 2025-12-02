package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/backbone81/membership/internal/encryption"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/utility"
	"github.com/backbone81/membership/pkg/membership"
)

var (
	verbosity                int
	maxDatagramLengthSend    int
	maxDatagramLengthReceive int
	bindAddress              string
	advertiseAddress         string

	protocolPeriod    time.Duration
	directPingTimeout time.Duration
	members           []string
	encryptionKeys    []string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "demo",
	Short:        "Demonstrates the use of the membership library.",
	Long:         `Demonstrates the use of the membership library.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		logger, zapLogger, err := utility.CreateLogger(verbosity)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		logger.Info("Application startup")

		resolveAdvertiseAddress, err := membership.ResolveAdvertiseAddress(advertiseAddress, bindAddress)
		if err != nil {
			return err
		}
		logger.Info(
			"Advertise address",
			"address", advertiseAddress,
			"ip", resolveAdvertiseAddress.IP(),
			"port", resolveAdvertiseAddress.Port(),
		)

		resolvedBootstrapMembers, err := membership.ResolveBootstrapMembers(members)
		if err != nil {
			return err
		}
		for i := range members {
			logger.Info(
				"Resolved member",
				"member", members[i],
				"ip", resolvedBootstrapMembers[i].IP(),
				"port", resolvedBootstrapMembers[i].Port(),
			)
		}

		options := []membership.Option{
			membership.WithLogger(logger),
			membership.WithProtocolPeriod(protocolPeriod),
			membership.WithBootstrapMembers(resolvedBootstrapMembers),
			membership.WithAdvertisedAddress(resolveAdvertiseAddress),
			membership.WithBindAddress(bindAddress),
			membership.WithMaxDatagramLengthSend(maxDatagramLengthSend),
			membership.WithMaxDatagramLengthReceive(maxDatagramLengthReceive),
		}
		for _, key := range encryptionKeys {
			typedKey, err := encryption.ParseKeyFromHexString(key)
			if err != nil {
				return err
			}
			options = append(options, membership.WithEncryptionKey(typedKey))
		}
		membershipList, err := membership.NewList(options...)
		if err != nil {
			return err
		}

		if err := membershipList.Startup(); err != nil {
			return err
		}

		logger.Info("Application running")
		<-ctx.Done()

		logger.Info("Application shutdown")
		if err := membershipList.Shutdown(); err != nil {
			return err
		}
		return nil
	},
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

	rootCmd.PersistentFlags().StringVar(
		&bindAddress,
		"bind-address",
		":3000",
		`The local address to bind to and accept incoming network messages.`,
	)
	rootCmd.PersistentFlags().StringVar(
		&advertiseAddress,
		"advertise-address",
		"",
		`The address which is used to advertise to other members.
This can be ip:port or host:port. The host will be resolved on startup. 
The port should match the port used for bind-address.
If left empty, the ip address of the host will be auto-detected.`,
	)

	rootCmd.PersistentFlags().DurationVar(
		&protocolPeriod,
		"protocol-period",
		1*time.Second,
		`The duration of a full protocol period with direct and indirect pings.
Any member which did not respond within that time is marked as suspect.
This should be at least three times the usual round-trip time between members.`,
	)
	rootCmd.PersistentFlags().DurationVar(
		&directPingTimeout,
		"direct-ping-timeout",
		100*time.Millisecond,
		`The duration after which an indirect ping is initiated.
This should be the usual round-trip time between members.`,
	)

	rootCmd.PersistentFlags().StringArrayVar(
		&members,
		"member",
		nil,
		`Other known member to connect to. Should be ip:port or host:port.
Hostname will be resolved to ip address on startup.
Can be specified multiple times to configure several members.`,
	)
	rootCmd.PersistentFlags().StringArrayVar(
		&encryptionKeys,
		"encryption-key",
		nil,
		`The hex encoded encryption key to use for encrypting and decrypting network messages. Can be given
multiple times for multiple encryption keys. The first key is always used for encryption, all keys are used in turn
until decryption succeeds.`,
	)
}
