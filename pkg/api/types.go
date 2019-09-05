package api

var (
	DefaultMachineImage = MachineImage{Name: "ubuntu", Version: "19.04"}
)

const (
	DefaultFirewallImage        = "firewall-1"
	DefaultLoadBalancerProvider = "metallb"
	DefaultVolumeType           = "storage_1"
	DefaultVolumeSize           = "200Gi"
)

type (
	// ShootCreateRequest is used to create new Metal Shoot cluster
	ShootCreateRequest struct {
		// Name is a human-readable name of the cluster
		Name string
		// Description is a human-readable description of what the cluster is used for.
		// +optional
		Description *string
		// CreatedBy is a subject representing a user name, an email address, or any other identifier of a user
		// who created the project.
		CreatedBy string
		// ProjectID is the metal projec in which the shoot will be placed
		ProjectID string
		// Owner is a subject representing a user name, an email address, or any other identifier of a user owning
		// the project.
		Owner string
		// Tenant is a the customer who owns this cluster.
		Tenant string
		// Purpose is a human-readable explanation of the project's purpose.
		// +optional
		Purpose *string
		// LoadBalancerProvider is the name of the load balancer provider in the Metal environment.
		LoadBalancerProvider string
		// MachineImage holds information about the machine image to use for all workers.
		// It will default to the first image stated in the referenced CloudProfile if no
		// value has been provided.
		// +optional
		MachineImage MachineImage
		// FirewallImage is the image of the firewall to use
		FirewallImage string
		// FirewallSize is the size of the firewall machine
		FirewallSize string
		// NodeNetwork is the network cidr where all machines/nodes will have their private network
		NodeNetwork string
		// AdditionalNetworks holds information about the Kubernetes and infrastructure networks except Nodenetwork.
		AdditionalNetworks []string
		// Workers is a list of worker groups.
		Workers []Worker
		// Zones is a list of availability zones to deploy the Shoot cluster to.
		Zones []string
		// Mainenance defines what and when a update of Worker nodes or Kubernetes should happen
		Maintenance Maintenance
		// Kubernetes defines k8s specific parameters
		Kubernetes Kubernetes
	}

	// Kubernetes carries kubernetes specific configuration options
	Kubernetes struct {
		// AllowPrivilegedContainers indicates whether privileged containers are allowed in the Shoot (default: true).
		AllowPrivilegedContainers bool
		// Version is the semantic Kubernetes version to use for the Shoot cluster.
		Version string
	}
	// Worker is the base definition of a worker group.
	Worker struct {
		// Name is the name of the worker group.
		Name string
		// MachineType is the machine type of the worker group.
		MachineType string
		// AutoScalerMin is the minimum number of VMs to create.
		AutoScalerMin int
		// AutoScalerMin is the maximum number of VMs to create.
		AutoScalerMax int
		// MaxSurge is maximum number of VMs that are created during an update.
		// +optional
		MaxSurge int
		// MaxUnavailable is the maximum number of VMs that can be unavailable during an update.
		// +optional
		MaxUnavailable int
		// VolumeType is the type of the root volumes.
		VolumeType string
		// VolumeSize is the size of the root volume.
		VolumeSize string
	}

	// MachineImage defines the name and the version of the machine image in any environment.
	MachineImage struct {
		Name    string
		Version string
	}
	// Maintenance contains information about the time window for maintenance operations and which
	// operations should be performed.
	Maintenance struct {
		// AutoUpdate contains information about which constraints should be automatically updated.
		// +optional
		AutoUpdate *MaintenanceAutoUpdate
		// TimeWindow contains information about the time window for maintenance operations.
		// +optional
		TimeWindow *MaintenanceTimeWindow
	}

	// MaintenanceAutoUpdate contains information about which constraints should be automatically updated.
	MaintenanceAutoUpdate struct {
		// KubernetesVersion indicates whether the patch Kubernetes version may be automatically updated.
		KubernetesVersion bool
		// MachineImage indicates whether the worker node should be updated automatically, e.g. replaced with a new machine.
		MachineImage bool
	}
	// MaintenanceTimeWindow contains information about the time window for maintenance operations.
	MaintenanceTimeWindow struct {
		// Begin is the beginning of the time window in the format HHMMSS+ZONE, e.g. "220000+0100".
		// If not present, a random value will be computed.
		Begin string
		// End is the end of the time window in the format HHMMSS+ZONE, e.g. "220000+0100".
		// If not present, the value will be computed based on the "Begin" value.
		End string
	}

	// ShootConstraints are configured in the Seed as constraint for a new Shoot
	ShootConstraints struct {
		KubernetesVersions []string
		Partitions         []string
	}
)
