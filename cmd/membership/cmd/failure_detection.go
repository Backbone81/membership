package cmd

import (
	"github.com/backbone81/membership/cmd/membership/cmd/failuredetection"
	"github.com/backbone81/membership/internal/utility"
	"github.com/spf13/cobra"
)

var (
	failureDetectionMinMemberCount int
	failureDetectionLinearCutoff   int
	failureDetectionMaxMemberCount int
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

		return failuredetection.Simulate(failureDetectionMinMemberCount, failureDetectionLinearCutoff, failureDetectionMaxMemberCount, logger)
	},
}

func init() {
	rootCmd.AddCommand(failureDetectionCmd)

	failureDetectionCmd.PersistentFlags().IntVar(
		&failureDetectionMinMemberCount,
		"min-member-count",
		2,
		"The minimum member count to simulate.",
	)
	failureDetectionCmd.PersistentFlags().IntVar(
		&failureDetectionLinearCutoff,
		"linear-cutoff",
		8,
		"Member counts increase linear between min-member-count and linear-cutoff. After linear-cutoff, member counts are doubled.",
	)
	failureDetectionCmd.PersistentFlags().IntVar(
		&failureDetectionMaxMemberCount,
		"max-member-count",
		512,
		"The maximum member count to simulate.",
	)
}
