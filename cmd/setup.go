package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ayildirim21/numaflow-perfman/setup"
	"github.com/ayildirim21/numaflow-perfman/setup/portforward"
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
			numaflowChart := setup.ChartRelease{
				ChartName:   "numaflow",
				ReleaseName: "perfman-numaflow",
				RepoUrl:     "https://numaproj.io/helm-charts",
				Namespace:   util.NumaflowNamespace,
				Values:      nil,
			}
			if err := numaflowChart.InstallOrUpgradeRelease(kubeClient, log); err != nil {
				return fmt.Errorf("unable to install numaflow: %w", err)
			}
		}

		// Optionally install ISB service
		if cmd.Flag("jetstream").Changed {
			isbGvro := setup.GVRObject{
				Group:     "numaflow.numaproj.io",
				Version:   "v1alpha1",
				Resource:  "interstepbufferservices",
				Namespace: util.DefaultNamespace,
			}
			if err := isbGvro.CreateResource("setup/isbvc.yaml", dynamicClient, log); err != nil {
				return fmt.Errorf("failed to create jetsream-isbvc: %w", err)
			}
		}

		// Install prometheus operator
		kubePrometheusChart := setup.ChartRelease{
			ChartName:   "kube-prometheus",
			ReleaseName: "perfman-kube-prometheus",
			RepoUrl:     "https://charts.bitnami.com/bitnami",
			Namespace:   util.DefaultNamespace,
			Values:      nil,
		}
		if err := kubePrometheusChart.InstallOrUpgradeRelease(kubeClient, log); err != nil {
			return fmt.Errorf("failed to install prometheus operator: %w", err)
		}

		// Install service monitors
		svGvro := setup.GVRObject{
			Group:     "monitoring.coreos.com",
			Version:   "v1",
			Resource:  "servicemonitors",
			Namespace: util.DefaultNamespace,
		}
		// TODO: check if service monitors exist before applying them
		if err := svGvro.CreateResource("setup/pipeline-metrics.yaml", dynamicClient, log); err != nil {
			return fmt.Errorf("failed to create service monitor for pipeline metrics: %w", err)
		}

		if err := svGvro.CreateResource("setup/isbvc-jetstream-metrics.yaml", dynamicClient, log); err != nil {
			return fmt.Errorf("failed to create service monitor for jetstream metrics: %w", err)
		}

		// Port forward prometheus operator to localhost:9090, so that it can be used as a source in Grafana dashboard
		options := []*portforward.Option{
			{
				LocalPort:   9090,
				RemotePort:  9090,
				ServiceName: "perfman-kube-prometheus-prometheus",
				Source:      "svc/perfman-kube-prometheus-prometheus",
				Namespace:   util.DefaultNamespace,
			},
		}

		ret, err := portforward.Forwarders(context.TODO(), options, config, kubeClient, log)
		if err != nil {
			return fmt.Errorf("failed to portforward %s: %w", options[0].Source, err)
		}

		defer ret.Close()

		ports, err := ret.Ready()
		if err != nil {
			return fmt.Errorf("failed to get ports: %w", err)
		}

		log.Info("successfully port forwarding prometheus operator to localhost:9090", zap.Any("ports", ports))

		ret.Wait()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVarP(&Numaflow, "numaflow", "n", false, "Install/upgrade the numaflow system")
	setupCmd.Flags().BoolVarP(&Jetstream, "jetstream", "j", false, "Install jetsream as the InterStepBuffer service")
}
