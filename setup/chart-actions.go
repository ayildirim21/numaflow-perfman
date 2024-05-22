package setup

import (
	"context"
	"errors"
	"fmt"
	logger "log"
	"os"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func InstallOrUpgradeRelease(releaseName string, repoUrl string, chartName string, values map[string]interface{}, targetNamespace string, kubeClient *kubernetes.Clientset, log *zap.Logger) error {
	nsName := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespace,
		},
	}

	if _, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), targetNamespace, metav1.GetOptions{}); err != nil {
		// if namespace does not exist, create it
		if _, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsName, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create namespace %s: %w", targetNamespace, err)
		}
	}

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), targetNamespace, os.Getenv("HELM_DRIVER"), logger.Printf); err != nil {
		return fmt.Errorf("failed to initialize actionConfig: %w", err)
	}

	chartPathOptions := action.ChartPathOptions{
		RepoURL: repoUrl,
	}

	c, err := getChart(chartPathOptions, chartName, settings)
	if err != nil {
		return fmt.Errorf("failed to get chart: %w", err)
	}

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); errors.Is(err, driver.ErrReleaseNotFound) {
		clientInstall := action.NewInstall(actionConfig)
		clientInstall.ReleaseName = releaseName
		clientInstall.Namespace = targetNamespace
		clientInstall.ChartPathOptions = chartPathOptions

		rel, err := clientInstall.Run(c, values)
		if err != nil {
			return fmt.Errorf("failed to install %s: %w", repoUrl, err)
		}

		log.Info("installed chart successfully", zap.String("release-name", rel.Name), zap.String("release-namespace", rel.Namespace))
	} else {
		clientUpgrade := action.NewUpgrade(actionConfig)
		clientUpgrade.Namespace = targetNamespace
		clientUpgrade.ChartPathOptions = chartPathOptions

		rel, err := clientUpgrade.Run(releaseName, c, values)
		if err != nil {
			return fmt.Errorf("failed to upgrade %s: %w", repoUrl, err)
		}

		log.Info("updated chart successfully", zap.String("release-name", rel.Name), zap.String("release-namespace", rel.Namespace))
	}

	return nil
}

func getChart(chartPathOption action.ChartPathOptions, chartName string, settings *cli.EnvSettings) (*chart.Chart, error) {
	chartPath, err := chartPathOption.LocateChart(chartName, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate %s: %w", chartName, err)
	}

	c, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", chartName, err)
	}

	return c, nil
}
