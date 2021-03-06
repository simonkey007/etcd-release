package helpers

import (
	"github.com/go-yaml/yaml"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	"github.com/pivotal-cf-experimental/destiny/etcd"
)

func DeploymentVMs(boshClient bosh.Client, deploymentName string) ([]bosh.VM, error) {
	vms, err := boshClient.DeploymentVMs(deploymentName)
	if err != nil {
		return nil, err
	}

	for index := range vms {
		vms[index].IPs = nil
	}

	return vms, nil
}

func GetVMIPs(boshClient bosh.Client, deploymentName, jobName string) ([]string, error) {
	vms, err := boshClient.DeploymentVMs(deploymentName)
	if err != nil {
		return []string{}, err
	}

	for _, vm := range vms {
		if vm.JobName == jobName {
			return vm.IPs, nil
		}
	}

	return []string{}, nil
}

func GetVMsFromManifest(manifest etcd.Manifest) []bosh.VM {
	var vms []bosh.VM

	for _, job := range manifest.Jobs {
		for i := 0; i < job.Instances; i++ {
			vms = append(vms, bosh.VM{JobName: job.Name, Index: i, State: "running"})
		}
	}

	return vms
}

func GetVMsFromManifestV2(manifest string) []bosh.VM {
	var vms []bosh.VM

	instanceGroups, err := etcd.InstanceGroups(manifest)
	if err != nil {
		panic(err)
	}

	for _, ig := range instanceGroups {
		for i := 0; i < ig.Instances; i++ {
			vms = append(vms, bosh.VM{JobName: ig.Name, Index: i, State: "running"})
		}
	}

	return vms
}

func GetNonErrandVMsFromRawManifest(rawManifest []byte) ([]bosh.VM, error) {
	var vms []bosh.VM

	var manifest Manifest
	err := yaml.Unmarshal(rawManifest, &manifest)
	if err != nil {
		return nil, err
	}

	for _, job := range manifest.Jobs {
		for i := 0; i < job.Instances; i++ {
			if job.Lifecycle != "errand" {
				vms = append(vms, bosh.VM{JobName: job.Name, Index: i, State: "running"})
			}
		}
	}

	return vms, nil
}

func GetVMIDByIndices(boshClient bosh.Client, deploymentName, jobName string, indices []int) ([]string, error) {
	vms, err := boshClient.DeploymentVMs(deploymentName)
	if err != nil {
		return []string{}, err
	}

	var vmIDs []string
	for _, vm := range vms {
		if vm.JobName == jobName {
			for _, index := range indices {
				if index == vm.Index {
					vmIDs = append(vmIDs, vm.ID)
				}
			}
		}
	}

	return vmIDs, nil
}

type Manifest struct {
	Name          interface{}            `yaml:"name"`
	DirectorUUID  string                 `yaml:"director_uuid"`
	Releases      interface{}            `yaml:"releases"`
	Jobs          []Job                  `yaml:"jobs"`
	Compilation   interface{}            `yaml:"compilation"`
	Networks      interface{}            `yaml:"networks"`
	Properties    map[string]interface{} `yaml:"properties"`
	ResourcePools interface{}            `yaml:"resource_pools"`
	Update        interface{}            `yaml:"update"`
	DiskPools     interface{}            `yaml:"disk_pools,omitempty"`
}

type Job struct {
	DefaultNetworks    []DefaultNetwork `yaml:"default_networks,omitempty"`
	Name               string           `yaml:"name"`
	Instances          int              `yaml:"instances"`
	PersistentDisk     *int             `yaml:"persistent_disk,omitempty"`
	PersistentDiskPool string           `yaml:"persistent_disk_pool,omitempty"`
	ResourcePool       string           `yaml:"resource_pool"`
	Networks           []Network        `yaml:"networks"`
	Update             *Update          `yaml:"update,omitempty"`
	Properties         *JobProperties   `yaml:"properties,omitempty"`
	Lifecycle          string           `yaml:"lifecycle,omitempty"`
	Templates          []Template       `yaml:"templates"`
}

type JobProperties struct {
	Consul            *PropertiesConsul `yaml:"consul,omitempty"`
	MetronAgent       interface{}       `yaml:"metron_agent,omitempty"`
	Router            interface{}       `yaml:"router,omitempty"`
	HAProxy           interface{}       `yaml:"ha_proxy,omitempty"`
	RouteRegistrar    interface{}       `yaml:"route_registrar,omitempty"`
	UAA               interface{}       `yaml:"uaa,omitempty"`
	NFSServer         interface{}       `yaml:"nfs_server,omitempty"`
	DEANext           interface{}       `yaml:"dea_next,omitempty"`
	Doppler           interface{}       `yaml:"doppler,omitempty"`
	TrafficController interface{}       `yaml:"traffic_controller,omitempty"`
	Diego             interface{}       `yaml:"diego,omitempty"`
	Loggregator       interface{}       `yaml:"loggregator,omitempty"`
}

type Template struct {
	Name     string      `yaml:"name"`
	Release  string      `yaml:"release"`
	Consumes interface{} `yaml:"consumes,omitempty"`
}

type DefaultNetwork struct {
	Name string
}

type Network struct {
	Name      string    `yaml:"name"`
	StaticIPs *[]string `yaml:"static_ips,omitempty"`
}

type Update struct {
	MaxInFlight int  `yaml:"max_in_flight,omitempty"`
	Serial      bool `yaml:"serial,omitempty"`
}
