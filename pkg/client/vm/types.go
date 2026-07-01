package vm

type VMInterface struct {
	IP           string `json:"ip"`
	MAC          string `json:"mac"`
	NetworkID    string `json:"network_id"`
	NetworkName  string `json:"network_name"`
	SubnetID     string `json:"subnet_id"`
	SubnetName   string `json:"subnet_name"`
	DeviceType   string `json:"device_type"`
	PrimaryMAC   string `json:"primary_mac"`
	Protected    bool   `json:"protected"`
	QueueNumbers int    `json:"queue_numbers"`
}

type VMDisk struct {
	Name          string                 `json:"name"`
	DiskType      string                 `json:"disk_type"`
	Size          int                    `json:"size"`
	Used          int                    `json:"used"`
	Encrypt       bool                   `json:"encrypt"`
	Autoscaling   bool                   `json:"autoscaling"`
	Scrub         bool                   `json:"scrub"`
	AdditionalDef map[string]interface{} `json:"additional_def,omitempty"`
}

type VMInstanceLight struct {
	VmUUID       string        `json:"vm_uuid"`
	VmFQDN       string        `json:"vm_fqdn"`
	Status       string        `json:"status"`
	Interfaces   []VMInterface `json:"interfaces"`
	OSTitle      string        `json:"os_title"`
	ServiceClass string        `json:"serviceclass"`
	Availability string        `json:"availability"`
	CPU          int           `json:"cpu"`
	RAM          int           `json:"ram"`
	StorageClass string        `json:"storageclass"`
	ServiceTitle string        `json:"service_title"`
}

type VMInstanceDetail struct {
	VmUUID       string        `json:"vm_uuid"`
	VmFQDN       string        `json:"vm_fqdn"`
	Status       string        `json:"status"`
	StatusInfo   string        `json:"status_info"`
	Interfaces   []VMInterface `json:"interfaces"`
	Disks        []VMDisk      `json:"disks"`
	OSTitle      string        `json:"os_title"`
	DomainUUID   string        `json:"domain_uuid"`
	DomainFQDN   string        `json:"domain_fqdn"`
	LdapUUID     string        `json:"ldap_uuid"`
	LdapFQDN     string        `json:"ldap_fqdn"`
	ProjectUUID  string        `json:"project_uuid"`
	ProjectTitle string        `json:"project_title"`
	TenantUUID   string        `json:"tenant_uuid"`
	TenantTitle  string        `json:"tenant_title"`
	Contract     bool          `json:"contract"`
	ServiceClass string        `json:"serviceclass"`
	Availability string        `json:"availability"`
	CPU          int           `json:"cpu"`
	RAM          int           `json:"ram"`
	StorageClass string        `json:"storageclass"`
	ServiceUUID  string        `json:"service_uuid"`
	ServiceTitle string        `json:"service_title"`
	ContactUUID  string        `json:"contact_uuid"`
	MailAddress  string        `json:"mailaddress"`
}
