package portforward

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
)

type Result struct {
	Close func()                                        // close the port forwarding
	Ready func() ([][]portforward.ForwardedPort, error) // block till the forwarding ready
	Wait  func()                                        // block and listen IOStreams close signal
}

type Option struct {
	LocalPort   int
	RemotePort  int
	Namespace   string
	PodName     string
	ServiceName string
	Source      string
}

type portForwardAPodRequest struct {
	RestConfig *rest.Config               // RestConfig is the kubernetes config
	Pod        v1.Pod                     // Pod is the selected pod for this port forwarding
	LocalPort  int                        // LocalPort is the local port that will be selected to expose the PodPort
	PodPort    int                        // PodPort is the target port for the pod
	Streams    genericiooptions.IOStreams // Steams configures where to write or read input from
	StopCh     <-chan struct{}            // StopCh is the channel used to manage the port forward lifecycle
	ReadyCh    chan struct{}              // ReadyCh communicates when the tunnel is ready to receive traffic
}

type PodOption struct {
	LocalPort int
	PodPort   int
	Pod       v1.Pod
}

type carry struct {
	StopCh  chan struct{}              // StopCh is the channel used to manage the port forward lifecycle
	ReadyCh chan struct{}              // ReadyCh communicates when the tunnel is ready to receive traffic
	PF      *portforward.PortForwarder // the instance of Portforwarder
}
