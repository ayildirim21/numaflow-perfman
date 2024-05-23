package setup

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

type GVRObject struct {
	Group     string
	Version   string
	Resource  string
	Namespace string
}

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

func (gvro *GVRObject) CreateResource(filename string, dynamicClient *dynamic.DynamicClient, logger *zap.Logger) error {
	obj, err := readYamlFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file for configuration information: %w", err)
	}

	gvr := schema.GroupVersionResource{Group: gvro.Group, Version: gvro.Version, Resource: gvro.Resource}
	result, err := dynamicClient.Resource(gvr).Namespace(gvro.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	logger.Info("Applied resource", zap.String("name", result.GetName()))
	return nil
}
