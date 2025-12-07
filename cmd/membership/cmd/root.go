package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/backbone81/membership/cmd/membership/cmd/failuredetection"
	"github.com/backbone81/membership/cmd/membership/cmd/failurepropagation"
	"github.com/backbone81/membership/cmd/membership/cmd/infection"
	"github.com/backbone81/membership/cmd/membership/cmd/joinpropagation"
	"github.com/backbone81/membership/cmd/membership/cmd/keygen"
	"github.com/backbone81/membership/cmd/membership/cmd/lossyjoin"
	"github.com/backbone81/membership/cmd/membership/cmd/statistics"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "membership",
	Short: "Command line tool supporting the membership Go library.",
	Long:  `Command line tool supporting the membership Go library.`,
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
	failuredetection.RegisterSubCommand(rootCmd)
	failurepropagation.RegisterSubCommand(rootCmd)
	infection.RegisterSubCommand(rootCmd)
	joinpropagation.RegisterSubCommand(rootCmd)
	keygen.RegisterSubCommand(rootCmd)
	lossyjoin.RegisterSubCommand(rootCmd)
	statistics.RegisterSubCommand(rootCmd)
}
