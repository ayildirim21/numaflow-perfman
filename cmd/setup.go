package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ayildirim21/numaflow-perfman/setup"
	"github.com/ayildirim21/numaflow-perfman/util"
)

var Numaflow bool
var Jetstream bool

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Deploy necessary services",
	Long:  "The setup command deploys Prometheus Operator as well as a couple Service Monitors onto the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		nonFlagArgs := cmd.Flags().Args()
		if len(nonFlagArgs) > 0 {
			return errors.New("this command doesn't accept args")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Optionally install numaflow
		if cmd.Flag("numaflow").Changed {
			releaseName := "perfman-numaflow"
			packagePath := "https://numaproj.io/helm-charts"
			chartName := "numaflow"
			if err := setup.InstallOrUpgradeRelease(releaseName, packagePath, chartName, nil, util.NumaflowNamespace, kubeClient, log); err != nil {
				return fmt.Errorf("unable to install numaflow: %w", err)
			}
		}

		// Optionally install ISB service
		if cmd.Flag("jetstream").Changed {
			isbGroup := "numaflow.numaproj.io"
			isbResource := "interstepbufferservices"
			isbPath := "setup/isbvc.yaml"
			if err := util.CreateResource(isbPath, dynamicClient, util.DefaultNamespace, isbGroup, "v1alpha1", isbResource, log); err != nil {
				return fmt.Errorf("unable to create jetsream-isbvc: %w", err)
			}
		}

		// Install service monitors
		svmGroup := "monitoring.coreos.com"
		svmResource := "servicemonitors"

		pipelineMetricsPath := "setup/pipeline-metrics.yaml"
		if err := util.CreateResource(pipelineMetricsPath, dynamicClient, util.DefaultNamespace, svmGroup, "v1", svmResource, log); err != nil {
			return fmt.Errorf("unable to create service monitor for pipeline metrics: %w", err)
		}
		jetstreamMetricsPath := "setup/isbvc-jetstream-metrics.yaml"
		if err := util.CreateResource(jetstreamMetricsPath, dynamicClient, util.DefaultNamespace, svmGroup, "v1", svmResource, log); err != nil {
			return fmt.Errorf("unable to create service monitor for jetstream metrics: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVarP(&Numaflow, "numaflow", "n", false, "Install/upgrade the numaflow system")
	setupCmd.Flags().BoolVarP(&Jetstream, "jetstream", "j", false, "Install jetsream as the InterStepBuffer service")
}
