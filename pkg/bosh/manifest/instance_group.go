package manifest

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"code.cloudfoundry.org/cf-operator/pkg/kube/apis"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

// InstanceGroup from BOSH deployment manifest.
type InstanceGroup struct {
	Name               string                 `json:"name"`
	Instances          int                    `json:"instances"`
	AZs                []string               `json:"azs"`
	Jobs               []Job                  `json:"jobs"`
	VMType             string                 `json:"vm_type,omitempty"`
	VMExtensions       []string               `json:"vm_extensions,omitempty"`
	VMResources        *VMResource            `json:"vm_resources"`
	Stemcell           string                 `json:"stemcell"`
	PersistentDisk     *int                   `json:"persistent_disk,omitempty"`
	PersistentDiskType string                 `json:"persistent_disk_type,omitempty"`
	Networks           []*Network             `json:"networks,omitempty"`
	Update             *Update                `json:"update,omitempty"`
	MigratedFrom       []*MigratedFrom        `json:"migrated_from,omitempty"`
	LifeCycle          InstanceGroupType      `json:"lifecycle,omitempty"`
	Properties         map[string]interface{} `json:"properties,omitempty"`
	Env                AgentEnv               `json:"env,omitempty"`
}

// NameSanitized returns the sanitized instance group name.
func (ig *InstanceGroup) NameSanitized() string {
	return names.Sanitize(ig.Name)
}

// ExtendedStatefulsetName constructs the ests name.
func (ig *InstanceGroup) ExtendedStatefulsetName(deploymentName string) string {
	ign := ig.NameSanitized()
	return fmt.Sprintf("%s-%s", deploymentName, ign)
}

// IndexedServiceName constructs an indexed service name. It's used to construct the other service
// names other than the headless service.
func (ig *InstanceGroup) IndexedServiceName(deploymentName string, index int) string {
	sn := serviceName(ig.Name, deploymentName, 53)
	return fmt.Sprintf("%s-%d", sn, index)
}

func (ig *InstanceGroup) jobInstances(
	deploymentName string,
	jobName string,
	spec JobSpec,
) []JobInstance {
	var jobsInstances []JobInstance
	for i := 0; i < ig.Instances; i++ {
		// TODO: Understand whether there are negative side-effects to using this
		// default or not.
		azs := []string{""}
		if len(ig.AZs) > 0 {
			azs = ig.AZs
		}

		for _, az := range azs {
			index := len(jobsInstances)
			address := ig.IndexedServiceName(deploymentName, index)
			name := fmt.Sprintf("%s-%s", ig.NameSanitized(), jobName)

			jobsInstances = append(jobsInstances, JobInstance{
				Address:   address,
				AZ:        az,
				Bootstrap: index == 0,
				Index:     index,
				Instance:  i,
				Name:      name,
				ID:        fmt.Sprintf("%s-%d-%s", ig.NameSanitized(), index, jobName),
			})
		}
	}
	return jobsInstances
}

// VMResource from BOSH deployment manifest.
type VMResource struct {
	CPU               int `json:"cpu"`
	RAM               int `json:"ram"`
	EphemeralDiskSize int `json:"ephemeral_disk_size"`
}

// Network from BOSH deployment manifest.
type Network struct {
	Name      string   `json:"name"`
	StaticIps []string `json:"static_ips,omitempty"`
	Default   []string `json:"default,omitempty"`
}

// Update from BOSH deployment manifest.
type Update struct {
	Canaries        int     `json:"canaries"`
	MaxInFlight     string  `json:"max_in_flight"`
	CanaryWatchTime string  `json:"canary_watch_time"`
	UpdateWatchTime string  `json:"update_watch_time"`
	Serial          bool    `json:"serial,omitempty"`
	VMStrategy      *string `json:"vm_strategy,omitempty"`
}

// MigratedFrom from BOSH deployment manifest.
type MigratedFrom struct {
	Name string `json:"name"`
	Az   string `json:"az,omitempty"`
}

// IPv6 from BOSH deployment manifest.
type IPv6 struct {
	Enable bool `json:"enable"`
}

// JobDir from BOSH deployment manifest.
type JobDir struct {
	Tmpfs     *bool  `json:"tmpfs,omitempty"`
	TmpfsSize string `json:"tmpfs_size,omitempty"`
}

var (
	// LabelDeploymentName is the name of a label for the deployment name.
	LabelDeploymentName = fmt.Sprintf("%s/deployment-name", apis.GroupName)
	// LabelInstanceGroupName is the name of a label for an instance group name.
	LabelInstanceGroupName = fmt.Sprintf("%s/instance-group-name", apis.GroupName)
	// LabelDeploymentVersion is the name of a label for the deployment's version.
	LabelDeploymentVersion = fmt.Sprintf("%s/deployment-version", apis.GroupName)
)

// AgentSettings from BOSH deployment manifest.
// These annotations and labels are added to kube resources.
// Affinity is added into the pod's definition.
type AgentSettings struct {
	Annotations                  map[string]string `json:"annotations,omitempty"`
	Labels                       map[string]string `json:"labels,omitempty"`
	Affinity                     *corev1.Affinity  `json:"affinity,omitempty"`
	DisableLogSidecar            bool              `json:"disable_log_sidecar,omitempty" yaml:"disable_log_sidecar,omitempty"`
	ServiceAccountName           string            `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	AutomountServiceAccountToken *bool             `json:"automountServiceAccountToken,omitempty" yaml:"automountServiceAccountToken,omitempty"`
}

// Set overrides labels and annotations with operator-owned metadata.
func (as *AgentSettings) Set(manifestName, igName, version string) {
	if as.Labels == nil {
		as.Labels = map[string]string{}
	}
	as.Labels[LabelDeploymentName] = manifestName
	as.Labels[LabelInstanceGroupName] = igName
	as.Labels[LabelDeploymentVersion] = version
}

// Agent from BOSH deployment manifest.
type Agent struct {
	Settings AgentSettings `json:"settings,omitempty"`
	Tmpfs    *bool         `json:"tmpfs,omitempty"`
}

// AgentEnvBoshConfig from BOSH deployment manifest.
type AgentEnvBoshConfig struct {
	Password              string  `json:"password,omitempty"`
	KeepRootPassword      string  `json:"keep_root_password,omitempty"`
	RemoveDevTools        *bool   `json:"remove_dev_tools,omitempty"`
	RemoveStaticLibraries *bool   `json:"remove_static_libraries,omitempty"`
	SwapSize              *int    `json:"swap_size,omitempty"`
	IPv6                  IPv6    `json:"ipv6,omitempty"`
	JobDir                *JobDir `json:"job_dir,omitempty"`
	Agent                 Agent   `json:"agent,omitempty"`
}

// AgentEnv from BOSH deployment manifest.
type AgentEnv struct {
	PersistentDiskFS           string             `json:"persistent_disk_fs,omitempty"`
	PersistentDiskMountOptions []string           `json:"persistent_disk_mount_options,omitempty"`
	AgentEnvBoshConfig         AgentEnvBoshConfig `json:"bosh,omitempty"`
}
