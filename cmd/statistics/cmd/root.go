package cmd

import (
	"os"

	"github.com/backbone81/membership/cmd/statistics/cmd/packetlossjoin"
	"github.com/backbone81/membership/internal/utility"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "statistics",
	Short:        "Outputs some statistics about membership clusters in different sizes.",
	Long:         `Outputs some statistics about membership clusters in different sizes.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utility.CreateLogger(0)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		//if err := firstdetection.FirstFailureDetection(logger); err != nil {
		//	return err
		//}
		//
		//if err := allfailuredetection.AllFailureDetection(logger); err != nil {
		//	return err
		//}

		if err := packetlossjoin.PacketLossJoin(logger); err != nil {
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cmd.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
