package util

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

func readYamlFile(filename string) (*unstructured.Unstructured, error) {
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

func CreateResource(filename string, dynamicClient *dynamic.DynamicClient, namespace string, group string, version string, resource string, logger *zap.Logger) error {
	obj, err := readYamlFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file for configuration information: %w", err)
	}

	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	result, err := dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	logger.Info("Applied resource", zap.String("name", result.GetName()))
	return nil
}
