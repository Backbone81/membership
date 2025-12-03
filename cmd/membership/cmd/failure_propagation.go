package cmd

import (
	"github.com/backbone81/membership/cmd/membership/cmd/failurepropagation"
	"github.com/backbone81/membership/internal/utility"
	"github.com/spf13/cobra"
)

var (
	failurePropagationMinMemberCount int
	failurePropagationLinearCutoff   int
	failurePropagationMaxMemberCount int
)

// failurePropagationCmd represents the allDetection command
var failurePropagationCmd = &cobra.Command{
	Use:   "failure-propagation",
	Short: "How long a cluster needs to propagate a failed member.",
	Long: `Simulates clusters of different sizes with one member failed.
Measures the number of protocol periods until all non-faulty members know about the failed member.
Note that this simulation does not execute the periodic full list sync which the default membership list would do.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utility.CreateLogger(0)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		return failurepropagation.Simulate(failurePropagationMinMemberCount, failurePropagationLinearCutoff, failurePropagationMaxMemberCount, logger)
	},
}

func init() {
	rootCmd.AddCommand(failurePropagationCmd)

	failurePropagationCmd.PersistentFlags().IntVar(
		&failurePropagationMinMemberCount,
		"min-member-count",
		2,
		"The minimum member count to simulate.",
	)
	failurePropagationCmd.PersistentFlags().IntVar(
		&failurePropagationLinearCutoff,
		"linear-cutoff",
		8,
		"Member counts increase linear between min-member-count and linear-cutoff. After linear-cutoff, member counts are doubled.",
	)
	failurePropagationCmd.PersistentFlags().IntVar(
		&failurePropagationMaxMemberCount,
		"max-member-count",
		512,
		"The maximum member count to simulate.",
	)
}
