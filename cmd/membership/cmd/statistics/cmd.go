package statistics

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/utility"
)

var (
	memberCount        int
	memberReliability  float64
	protocolPeriod     time.Duration
	networkReliability float64
	safetyFactor       float64
)

// statisticsCmd represents the allDetection command
var statisticsCmd = &cobra.Command{
	Use:          "statistics",
	Short:        "Displays some analytical statistics about clusters.",
	Long:         `Some aspects of a cluster can be calculated without simulation and compared to the simulation for plausibility.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

		logger.Info("Cluster configuration",
			"member-count", memberCount,
			"member-reliability", memberReliability,
			"protocol-period", protocolPeriod,
			"network-reliability", networkReliability,
			"safety-factor", safetyFactor,
		)
		logger.Info("Probability of a member being chosen as direct ping target (0 to 1)",
			"probability", utility.PingTargetProbability(memberCount, memberReliability),
		)
		logger.Info("Expected time between a member failing and being detected",
			"duration", utility.FailureDetectionDuration(protocolPeriod, memberCount, memberReliability),
		)
		logger.Info("Probability of a member falsely being detected as failed",
			"probability", utility.FailureDetectionFalsePositiveProbability(networkReliability, memberReliability),
		)
		logger.Info("Protocol period count for dissemination",
			"protocol-periods", int(math.Ceil(utility.DisseminationPeriods(safetyFactor, memberCount))),
		)
		return nil
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(statisticsCmd)

	statisticsCmd.PersistentFlags().IntVar(
		&memberCount,
		"member-count",
		512,
		"The member count of the cluster.",
	)
	statisticsCmd.PersistentFlags().Float64Var(
		&memberReliability,
		"member-reliability",
		1.0,
		"The probability a member is available as a value between 0 and 1.",
	)
	statisticsCmd.PersistentFlags().DurationVar(
		&protocolPeriod,
		"protocol-period",
		1.0*time.Second,
		"The duration of a protocol period.",
	)
	statisticsCmd.PersistentFlags().Float64Var(
		&networkReliability,
		"network-reliability",
		1.0,
		"The probability of a network message arriving at its target as a value between 0 and 1.",
	)
	statisticsCmd.PersistentFlags().Float64Var(
		&safetyFactor,
		"safety-factor",
		3.0,
		"The safety factor to apply for making sure that changes are gossiped enough times.",
	)
}
