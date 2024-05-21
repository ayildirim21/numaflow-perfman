package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

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
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	util.InitializeClients()
	logger = logging.CreateLogger()
}
