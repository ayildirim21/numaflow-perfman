package service_monitors

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

func readServiceMonitorFile(filename string) (*unstructured.Unstructured, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml file: %w", err)
	}

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(yamlFile, &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to unmasrhsal into object: %w", err)
	}

	return &obj, nil
}

func CreateServiceMonitor(filename string, logger *zap.Logger, dynamicClient *dynamic.DynamicClient) error {
	serviceMonitor, err := readServiceMonitorFile(filename)
	if err != nil {
		panic(err)
	}

	gvr := schema.GroupVersionResource{Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors"}
	result, err := dynamicClient.Resource(gvr).Namespace("default").Create(context.TODO(), serviceMonitor, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Service Monitor: %w", err)
	} else {
		logger.Info("Applied Service Monitor", zap.String("name", result.GetName()))
		return nil
	}
}
