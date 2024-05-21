package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	sv "github.com/ayildirim21/numaflow-perfman/service-monitors"
)

// initCmd represents the setup command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Deploy necessary services",
	Long:  "The init command deploys Prometheus Operator as well as a couple Service Monitors onto the cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sv.CreateServiceMonitor("service-monitors/pipeline-metrics.yaml", log, dynamicClient); err != nil {
			return fmt.Errorf("unable to create service monitor for pipeline metrics: %w", err)
		}

		if err := sv.CreateServiceMonitor("service-monitors/isbvc-jetstream-metrics.yaml", log, dynamicClient); err != nil {
			return fmt.Errorf("unable to create service monitor for jetstream metrics: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
