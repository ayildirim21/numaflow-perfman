package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/ayildirim21/numaflow-perfman/logging"
	"github.com/ayildirim21/numaflow-perfman/util"
)

var kubeClient *kubernetes.Clientset
var dynamicClient *dynamic.DynamicClient
var log *zap.SugaredLogger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "perfman",
	Short: "Numaflow performance testing framework",
	Long:  "Perfman is a command line utility for performance testing changes to the numaflow platform",
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	config, err := util.K8sRestConfig()
	if err != nil {
		panic(err)
	}

	kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	log = logging.CreateLogger()
}
