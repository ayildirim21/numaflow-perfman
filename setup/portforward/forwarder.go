package portforward

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var once sync.Once

func portForwardAPod(req *portForwardAPodRequest) (*portforward.PortForwarder, error) {
	targetURL, err := url.Parse(req.RestConfig.Host)
	if err != nil {
		return nil, err
	}

	targetURL.Path = path.Join(
		"api", "v1",
		"namespaces", req.Pod.Namespace,
		"pods", req.Pod.Name,
		"portforward",
	)

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, targetURL)
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	return fw, nil
}

func Forwarders(ctx context.Context, options []*Option, config *rest.Config, kubeClient *kubernetes.Clientset, log *zap.Logger) (*Result, error) {
	newOptions, errNewOptions := parseOptions(options)
	if errNewOptions != nil {
		return nil, errNewOptions
	}

	podOptions, errHandleOptions := handleOptions(ctx, newOptions, kubeClient, log)
	if errHandleOptions != nil {
		return nil, errHandleOptions
	}

	stream := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	carries := make([]*carry, len(podOptions))

	var g errgroup.Group

	for index, option := range podOptions {
		index2 := index
		stopCh := make(chan struct{}, 1)
		readyCh := make(chan struct{})

		req := &portForwardAPodRequest{
			RestConfig: config,
			Pod:        option.Pod,
			LocalPort:  option.LocalPort,
			PodPort:    option.PodPort,
			Streams:    stream,
			StopCh:     stopCh,
			ReadyCh:    readyCh,
		}
		g.Go(func() error {
			pf, errPortForward := portForwardAPod(req)
			if errPortForward != nil {
				return errPortForward
			}
			carries[index2] = &carry{StopCh: stopCh, ReadyCh: readyCh, PF: pf}
			return nil
		})
	}

	if errWait := g.Wait(); errWait != nil {
		return nil, errWait
	}

	ret := &Result{
		Close: func() {
			once.Do(func() {
				for _, c := range carries {
					close(c.StopCh)
				}
			})
		},
		Ready: func() ([][]portforward.ForwardedPort, error) {
			var pfs [][]portforward.ForwardedPort
			for _, c := range carries {
				<-c.ReadyCh
				ports, errGetPorts := c.PF.GetPorts()
				if errGetPorts != nil {
					return nil, errGetPorts
				}
				pfs = append(pfs, ports)
			}
			return pfs, nil
		},
	}

	ret.Wait = func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		log.Info("Terminating connection...")
		ret.Close()
	}

	go func() {
		<-ctx.Done()
		ret.Close()
	}()

	return ret, nil
}
