package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	sv "github.com/ayildirim21/numaflow-perfman/service-monitors"
)

// initCmd represents the setup command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Deploy necessary services",
	Long:  "The init command deploys Prometheus Operator as well as a couple Service Monitors onto the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := sv.CreateServiceMonitor("service-monitors/pipeline-metrics.yaml", log, dynamicClient); err != nil {
			log.Error("unable to create service monitor for pipeline metrics", zap.Error(err))
			os.Exit(1)
		}

		if err := sv.CreateServiceMonitor("service-monitors/isbvc-jetstream-metrics.yaml", log, dynamicClient); err != nil {
			log.Error("unable to create service monitor for jetstream metrics", zap.Error(err))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
