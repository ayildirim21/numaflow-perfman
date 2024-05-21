package cmd

import (
	"os"

	"go.uber.org/zap"

	"github.com/spf13/cobra"

	"github.com/ayildirim21/numaflow-perfman/logging"
	"github.com/ayildirim21/numaflow-perfman/util"
)

var logger *zap.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "perfman",
	Short: "Numaflow performance testing framework",
	Long:  "Perfman is a command line utility for performance testing changes to the numaflow platform",
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	logger = logging.CreateLogger()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()
}

func init() {
	util.InitializeClients()
}
