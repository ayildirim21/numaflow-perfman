package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayildirim21/numaflow-perfman/setup"
	"github.com/ayildirim21/numaflow-perfman/util"
)

var Numaflow bool

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Deploy necessary services",
	Long:  "The setup command deploys Prometheus Operator as well as a couple Service Monitors onto the cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Optionally install numaflow
		numaflowFlag := cmd.Flag("numaflow")
		if numaflowFlag.Changed {
			releaseName := "perfman-numaflow"
			packagePath := "https://numaproj.io/helm-charts"
			chartName := "numaflow"
			if err := setup.InstallOrUpgradeRelease(releaseName, packagePath, chartName, nil, util.NumaflowNamespace, kubeClient, log); err != nil {
				return fmt.Errorf("unable to install numaflow: %w", err)
			}
		}

		// Install service monitors
		pipelineMetricsPath := "setup/pipeline-metrics.yaml"
		if err := setup.CreateServiceMonitor(pipelineMetricsPath, dynamicClient, util.DefaultNamespace, log); err != nil {
			return fmt.Errorf("unable to create service monitor for pipeline metrics: %w", err)
		}
		jetstreamMetricsPath := "setup/isbvc-jetstream-metrics.yaml"
		if err := setup.CreateServiceMonitor(jetstreamMetricsPath, dynamicClient, util.DefaultNamespace, log); err != nil {
			return fmt.Errorf("unable to create service monitor for jetstream metrics: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVarP(&Numaflow, "numaflow", "n", false, "Install or upgrade the numaflow system")
}
