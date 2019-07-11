package main

/**
 * This code is based off of https://github.com/RadeonOpenCompute/k8s-device-plugin
 * A special thanks to the RadeonOpenCompute group for helping with simplifying the Device Plugin setup
**/

// Kubernetes (k8s) device plugin to enable registration of AMD GPU to a container cluster

import (
	"flag"
	"fmt"
	"log"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	dummyDeviceCount = 10000 // set dummy devices to be 10000, it is an arbirtary number, but must be larger than the amount of device SRIOV VF's
	rdmaDeviceDir    = "/dev/infiniband"
)

// Plugin is identical to DevicePluginServer interface of device plugin API.
type Plugin struct {
	stop chan bool
}

// Start is an optional interface that could be implemented by plugin.
// If case Start is implemented, it will be executed by Manager after
// plugin instantiation and before its registration to kubelet. This
// method could be used to prepare resources before they are offered
// to Kubernetes.
func (p *Plugin) Start() error {
	return nil
}

// Stop is an optional interface that could be implemented by plugin.
// If case Stop is implemented, it will be executed by Manager after the
// plugin is unregistered from kubelet. This method could be used to tear
// down resources.
func (p *Plugin) Stop() error {
	return nil
}

// GetDevicePluginOptions returns options to be communicated with Device
// Manager
func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

// PreStartContainer is expected to be called before each container start if indicated by plugin during registration phase.
// PreStartContainer allows kubelet to pass reinitialized devices to containers.
// PreStartContainer allows Device Plugin to run device specific operations on the Devices requested
func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	//shouldnt be called b/c prestart is not required as specified in GetDevicePluginOptions
	return &pluginapi.PreStartContainerResponse{}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devices := make([]*pluginapi.Device, dummyDeviceCount, dummyDeviceCount)

	for i := 0; i < dummyDeviceCount; i++ {
		devices[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("SRIOV-Device-%d", i),
			Health: pluginapi.Healthy,
		}
	}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: devices})

	for {
		select {
		case <-p.stop:
			//should never happend b/c heartbeat is never called in this app
			//just used to block the call, but if it is called in the future, shutdown all devices
			var health = pluginapi.Unhealthy
			for i := 0; i < dummyDeviceCount; i++ {
				devices[i].Health = health
			}
			s.Send(&pluginapi.ListAndWatchResponse{Devices: devices})
		}
	}
	// returning a value with this function will unregister the plugin from k8s
}

// Allocate is called during container creation so that the Device
// Plugin can run device specific operations and instruct Kubelet
// of the steps to make the Device available in the container
func (p *Plugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Println("Allocate: request contains: ", r)

	requestLen := len(r.GetContainerRequests())
	containerResponses := make([]*pluginapi.ContainerAllocateResponse, requestLen, requestLen)
	for i := 0; i < requestLen; i++ {
		deviceSpec := []*pluginapi.DeviceSpec{
			&pluginapi.DeviceSpec{
				HostPath:      rdmaDeviceDir,
				ContainerPath: rdmaDeviceDir,
				Permissions:   "rwm",
			},
		}
		containerResponses[i] = &pluginapi.ContainerAllocateResponse{
			Devices: deviceSpec,
		}
	}

	response := pluginapi.AllocateResponse{
		ContainerResponses: containerResponses,
	}
	log.Println("Allocate: response contains: ", response)
	return &response, nil
}

// Lister serves as an interface between imlementation and Manager machinery. User passes
// implementation of this interface to NewManager function. Manager will use it to obtain resource
// namespace, monitor available resources and instantate a new plugin for them.
type Lister struct{}

// GetResourceNamespace must return namespace (vendor ID) of implemented Lister. e.g. for
// resources in format "color.example.com/<color>" that would be "color.example.com".
func (l *Lister) GetResourceNamespace() string {
	return "rdma-sriov"
}

// Discover notifies manager with a list of currently available resources in its namespace.
// e.g. if "color.example.com/red" and "color.example.com/blue" are available in the system,
// it would pass PluginNameList{"red", "blue"} to given channel. In case list of
// resources is static, it would use the channel only once and then return. In case the list is
// dynamic, it could block and pass a new list each times resources changed. If blocking is
// used, it should check whether the channel is closed, i.e. Discover should stop.
func (l *Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	log.Printf("Discovered 1 socket for grpc\n")
	pluginListCh <- []string{"vf"}
}

// NewPlugin instantiates a plugin implementation. It is given the last name of the resource,
// e.g. for resource name "color.example.com/red" that would be "red". It must return valid
// implementation of a PluginInterface.
func (l *Lister) NewPlugin(resourceLastName string) dpm.PluginInterface {
	return &Plugin{}
}

func main() {
	// this is also needed to enable glog usage in dpm
	flag.Parse()

	//set the flag glog (used in dpm) for stderr to true b/c or image is a scratch
	//if not set to true, than it will write to a file, which scratch image does not have
	flag.Set("logtostderr", "true")

	log.Println("RDMA Dummy Device Plugin Starting Up...")
	l := Lister{}
	manager := dpm.NewManager(&l)
	manager.Run()
}
