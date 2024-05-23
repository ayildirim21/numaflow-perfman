package portforward

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ayildirim21/numaflow-perfman/util"
)

func parseSource(source string) (*Option, error) {
	list := strings.Split(source, "/")
	if len(list) != 2 {
		return nil, fmt.Errorf("invalid source: %v", source)
	}

	kind := list[0]
	name := list[1]

	if kind == "svc" || kind == "service" || kind == "services" {
		return &Option{ServiceName: name}, nil
	}

	if kind == "po" || kind == "pod" || kind == "pods" {
		return &Option{PodName: name}, nil
	}

	return nil, fmt.Errorf("invalid source: %v", source)
}

func parseOptions(options []*Option) ([]*Option, error) {
	var newOptions []*Option

	for _, o := range options {
		if o.Namespace == "" {
			o.Namespace = util.DefaultNamespace
		}
		if o.Source != "" {
			opt, err := parseSource(o.Source)
			if err != nil {
				return nil, err
			}
			if opt.ServiceName != "" {
				o.ServiceName = opt.ServiceName
			}
			if opt.PodName != "" {
				o.PodName = opt.PodName
			}
		}

		if o.PodName == "" && o.ServiceName == "" {
			return nil, fmt.Errorf("please provide a name for a pod or service")
		}

		newOptions = append(newOptions, o)
	}

	return newOptions, nil
}

func handleOptions(ctx context.Context, options []*Option, kubeClient *kubernetes.Clientset, log *zap.Logger) ([]*PodOption, error) {
	podOptions := make([]*PodOption, len(options))

	var g errgroup.Group

	for index, option := range options {
		option2 := option
		index2 := index
		g.Go(func() error {
			if option2.PodName != "" {
				pod, err := kubeClient.CoreV1().Pods(option2.Namespace).Get(ctx, option2.PodName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if pod == nil {
					return fmt.Errorf("no such pod: %v", option2.PodName)
				}

				podOptions[index2] = buildPodOption(option2, pod)
				return nil
			}

			svc, err := kubeClient.CoreV1().Services(option2.Namespace).Get(ctx, option2.ServiceName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if svc == nil {
				return fmt.Errorf("no such service: %+v", option2.ServiceName)
			}

			var labels []string
			for key, val := range svc.Spec.Selector {
				labels = append(labels, key+"="+val)
			}
			label := strings.Join(labels, ",")

			pods, err := kubeClient.CoreV1().Pods(option2.Namespace).List(ctx, metav1.ListOptions{LabelSelector: label, Limit: 1})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				return fmt.Errorf("no such pods for service %v", option2.ServiceName)
			}
			pod := pods.Items[0]

			log.Info("Forwarding service...", zap.String("service-name", option2.ServiceName), zap.String("pod-name", pod.Name))

			podOptions[index2] = buildPodOption(option2, &pod)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return podOptions, nil
}

func buildPodOption(option *Option, pod *v1.Pod) *PodOption {
	if option.RemotePort == 0 {
		option.RemotePort = int(pod.Spec.Containers[0].Ports[0].ContainerPort)
	}

	return &PodOption{
		LocalPort: option.LocalPort,
		PodPort:   option.RemotePort,
		Pod: v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		},
	}
}
