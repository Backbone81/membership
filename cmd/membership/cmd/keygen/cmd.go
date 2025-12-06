package keygen

import (
	"github.com/backbone81/membership/internal/encryption"
	"github.com/backbone81/membership/internal/utility"
	"github.com/spf13/cobra"
)

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:          "keygen",
	Short:        "Creates a random encryption key.",
	Long:         `Creates a random encryption key.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, zapLogger, err := utility.CreateLogger(0)
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		logger.Info("Random encryption key created", "key", encryption.NewRandomKey().String())
		return nil
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(keygenCmd)
}
