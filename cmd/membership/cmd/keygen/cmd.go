package keygen

import (
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"github.com/backbone81/membership/internal/encryption"
)

// keygenCmd represents the keygen command.
var keygenCmd = &cobra.Command{
	Use:          "keygen",
	Short:        "Creates a random encryption key.",
	Long:         `Creates a random encryption key.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

		logger.Info("Random encryption key created", "key", encryption.NewRandomKey().String())
		return nil
	},
}

func RegisterSubCommand(command *cobra.Command) {
	command.AddCommand(keygenCmd)
}
