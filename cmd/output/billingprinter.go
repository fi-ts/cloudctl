package output

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/spf13/viper"
)

type (
	// ProjectBillingTablePrinter print bills in a Table
	ProjectBillingTablePrinter struct {
		tablePrinter
	}

	// ClusterBillingTablePrinter print bills in a Table
	ClusterBillingTablePrinter struct {
		tablePrinter
	}
	// MachineBillingTablePrinter print bills in a Table
	MachineBillingTablePrinter struct {
		tablePrinter
	}
	// ProductOptionBillingTablePrinter print bills in a Table
	ProductOptionBillingTablePrinter struct {
		tablePrinter
	}
	// ContainerBillingTablePrinter print bills in a Table
	ContainerBillingTablePrinter struct {
		tablePrinter
	}
	// IPBillingTablePrinter print bills in a Table
	IPBillingTablePrinter struct {
		tablePrinter
	}
	// NetworkTrafficBillingTablePrinter print bills in a Table
	NetworkTrafficBillingTablePrinter struct {
		tablePrinter
	}
	// S3BillingTablePrinter print bills in a Table
	S3BillingTablePrinter struct {
		tablePrinter
	}
	// VolumeBillingTablePrinter print bills in a Table
	VolumeBillingTablePrinter struct {
		tablePrinter
	}
	// PostgresBillingTablePrinter print bills in a Table
	PostgresBillingTablePrinter struct {
		tablePrinter
	}
)

// Print a cluster usage as table
func (s ProjectBillingTablePrinter) Print(data []*models.V1ProjectInfoResponse) {
	s.wideHeader = []string{"Tenant", "ProjectID"}
	s.shortHeader = s.wideHeader
	if s.order == "" {
		s.order = "tenant,project"
	}
	s.Order(data)
	for _, u := range data {
		var tenant string
		if u.Tenantid != nil {
			tenant = *u.Tenantid
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}

		wide := []string{
			tenant,
			projectID,
		}
		short := wide

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	s.render()
}

// Print a cluster usage as table
func (s ClusterBillingTablePrinter) Print(data *models.V1ClusterUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "ClusterStart", "ClusterEnd", "Lifetime", "Group Avg", "Workers"}
	s.shortHeader = []string{"Tenant", "ProjectID", "Partition", "ClusterID", "ClusterName", "ClusterStart", "ClusterEnd", "Lifetime", "Group Avg", "Workers"}
	if s.order == "" {
		s.order = "tenant,project,partition,name,id"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var clusterStart string
		if u.Clusterstart != nil {
			clusterStart = u.Clusterstart.String()
		}
		var clusterEnd string
		if u.Clusterend != nil {
			clusterEnd = u.Clusterend.String()
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		var averageGroups string
		if u.Averageworkergroups != nil {
			if s, err := strconv.ParseFloat(*u.Averageworkergroups, 64); err == nil {
				averageGroups = fmt.Sprintf("%g", s)
			}

		}
		workers := "-"
		var workerCount int64
		for _, w := range u.Workergroups {
			if w.Machinecount == nil {
				continue
			}
			workerCount += *w.Machinecount
		}
		if workerCount > 0 {
			workerPlural := ""
			if len(u.Workergroups) > 1 {
				workerPlural = "s"
			}
			workers = fmt.Sprintf("%s (%d Group%s)", strconv.FormatInt(workerCount, 10), len(u.Workergroups), workerPlural)
		}

		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			clusterID,
			clusterName,
			clusterStart,
			clusterEnd,
			humanizeDuration(lifetime),
			averageGroups,
			workers,
		}
		short := []string{
			tenant,
			projectID,
			partition,
			clusterID,
			clusterName,
			clusterStart,
			clusterEnd,
			humanizeDuration(lifetime),
			averageGroups,
			workers,
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)) + lifetimeCosts(*data.Accumulatedusage.Lifetime), "", "",
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a machine usage as table
func (s MachineBillingTablePrinter) Print(data *models.V1MachineUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "Size", "MachineID", "MachineName", "ClusterID", "MachineStart", "Lifetime"}
	s.shortHeader = s.wideHeader

	if s.order == "" {
		s.order = "tenant,project,partition,name,id"
	}
	s.Order(data.Usage)

	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var machineID string
		if u.Machineid != nil {
			machineID = *u.Machineid
		}
		var machineName string
		if u.Machinename != nil {
			machineName = *u.Machinename
		}
		var sizeid string
		if u.Sizeid != nil {
			sizeid = *u.Sizeid
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var machineStart string
		if u.Machinestart != nil {
			machineStart = u.Machinestart.String()
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}

		row := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			sizeid,
			machineID,
			machineName,
			clusterID,
			machineStart,
			humanizeDuration(lifetime),
		}

		s.addWideData(row, data)
		s.addShortData(row, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)) + lifetimeCosts(*data.Accumulatedusage.Lifetime), "", "",
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a product option usage as table
func (s ProductOptionBillingTablePrinter) Print(data *models.V1ProductOptionUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Option", "ClusterID", "ClusterName", "Lifetime"}
	s.shortHeader = s.wideHeader

	if s.order == "" {
		s.order = "tenant,project,partition,name,id"
	}
	s.Order(data.Usage)

	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var option string
		if u.ID != nil {
			option = *u.ID
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}

		row := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			option,
			clusterID,
			clusterName,
			humanizeDuration(lifetime),
		}

		s.addWideData(row, data)
		s.addShortData(row, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)) + lifetimeCosts(*data.Accumulatedusage.Lifetime), "", "",
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a volume usage as table
func (s VolumeBillingTablePrinter) Print(data *models.V1VolumeUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "Start", "End", "UUID", "Name", "Type", "CapacitySeconds (Gi * h)", "Lifetime"}
	s.shortHeader = []string{"Tenant", "ProjectID", "Partition", "ClusterName", "UUID", "Name", "Type", "CapacitySeconds (Gi * h)", "Lifetime"}
	if s.order == "" {
		s.order = "tenant,project,partition,cluster,name"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var start string
		if u.Start != nil {
			start = u.Start.String()
		}
		var end string
		if u.End != nil {
			end = u.End.String()
		}
		var name string
		if u.Name != nil {
			name = *u.Name
		}
		var uuid string
		if u.UUID != nil {
			uuid = *u.UUID
		}
		var volumeType string
		if u.Type != nil {
			volumeType = *u.Type
		}
		var capacity string
		if u.Capacityseconds != nil {
			capacity = humanizeMemory(*u.Capacityseconds)
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			clusterID,
			clusterName,
			start,
			end,
			uuid,
			name,
			volumeType,
			capacity,
			humanizeDuration(lifetime),
		}
		short := []string{
			tenant,
			projectID,
			partition,
			clusterName,
			uuid,
			name,
			volumeType,
			capacity,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	var capacity string
	if data.Accumulatedusage.Capacityseconds != nil {
		capacity = humanizeMemory(*data.Accumulatedusage.Capacityseconds) + storageCosts(*data.Accumulatedusage.Capacityseconds)
	}
	var lifetime string
	if data.Accumulatedusage.Lifetime != nil {
		lifetime = humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime))
	}
	footer := []string{"Total",
		capacity,
		lifetime,
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a cluster usage as table
func (s IPBillingTablePrinter) Print(data *models.V1IPUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "IP", "Start", "End", "Lifetime"}
	s.shortHeader = []string{"Tenant", "ProjectID", "IP", "Start", "End", "Lifetime"}
	if s.order == "" {
		s.order = "tenant,project,ip"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var ip string
		if u.IP != nil {
			ip = *u.IP
		}
		var start string
		if u.Start != nil {
			start = u.Start.String()
		}
		var end string
		if u.End != nil {
			end = u.End.String()
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			ip,
			start,
			end,
			humanizeDuration(lifetime),
		}
		short := []string{
			tenant,
			projectID,
			ip,
			start,
			end,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)) + lifetimeCosts(*data.Accumulatedusage.Lifetime),
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a volume usage as table
func (s NetworkTrafficBillingTablePrinter) Print(data *models.V1NetworkUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "Device", "In (Gi)", "Out (Gi)", "Total (Gi)", "Lifetime"}
	s.shortHeader = []string{"Tenant", "ProjectID", "Partition", "ClusterName", "Device", "In (Gi)", "Out (Gi)", "Total (Gi)", "Lifetime"}
	if s.order == "" {
		s.order = "tenant,project,partition,cluster,device"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var device string
		if u.Device != nil {
			device = *u.Device
		}
		var in string
		if u.In != nil {
			in = humanizeBytesToGi(*u.In)
		}
		var out string
		if u.Out != nil {
			out = humanizeBytesToGi(*u.Out)
		}
		var total string
		if u.Total != nil {
			total = humanizeBytesToGi(*u.Total)
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			clusterID,
			clusterName,
			device,
			in,
			out,
			total,
			humanizeDuration(lifetime),
		}
		short := []string{
			tenant,
			projectID,
			partition,
			clusterName,
			device,
			in,
			out,
			total,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	var in string
	if data.Accumulatedusage.In != nil {
		in = humanizeBytesToGi(*data.Accumulatedusage.In) + giCosts(*data.Accumulatedusage.In, viper.GetFloat64("costs-incoming-network-traffic-gi"))
	}
	var out string
	if data.Accumulatedusage.Out != nil {
		out = humanizeBytesToGi(*data.Accumulatedusage.Out) + giCosts(*data.Accumulatedusage.Out, viper.GetFloat64("costs-outgoing-network-traffic-gi"))
	}
	var total string
	if data.Accumulatedusage.Total != nil {
		total = humanizeBytesToGi(*data.Accumulatedusage.Total) + giCosts(*data.Accumulatedusage.Total, viper.GetFloat64("costs-total-network-traffic-gi"))
	}
	var lifetime string
	if data.Accumulatedusage.Lifetime != nil {
		lifetime = humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime))
	}
	footer := []string{"Total",
		in,
		out,
		total,
		lifetime,
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a s3 usage as table
func (s S3BillingTablePrinter) Print(data *models.V1S3UsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "User", "Bucket Name", "Bucket ID", "Start", "End", "Objects", "StorageSeconds (Gi * h)", "Lifetime"}
	s.shortHeader = []string{"Tenant", "ProjectID", "Partition", "User", "Bucket Name", "Bucket ID", "Objects", "StorageSeconds (Gi * h)", "Lifetime"}
	if s.order == "" {
		s.order = "tenant,project,partition,user,bucket,bucket_id"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var user string
		if u.User != nil {
			user = *u.User
		}
		var bucketName string
		if u.Bucketname != nil {
			bucketName = *u.Bucketname
		}
		var bucketID string
		if u.Bucketid != nil {
			bucketID = *u.Bucketid
		}
		var start string
		if u.Start != nil {
			start = u.Start.String()
		}
		var end string
		if u.End != nil {
			end = u.End.String()
		}
		var objects string
		if u.Currentnumberofobjects != nil {
			objects = *u.Currentnumberofobjects
		}
		var storage string
		if u.Storageseconds != nil {
			storage = humanizeMemory(*u.Storageseconds)
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			user,
			bucketName,
			bucketID,
			start,
			end,
			objects,
			storage,
			humanizeDuration(lifetime),
		}
		short := []string{
			tenant,
			projectID,
			partition,
			user,
			bucketName,
			bucketID,
			objects,
			storage,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	objects := "0"
	if data.Accumulatedusage.Currentnumberofobjects != nil {
		objects = *data.Accumulatedusage.Currentnumberofobjects
	}
	var storage string
	if data.Accumulatedusage.Storageseconds != nil {
		storage = humanizeMemory(*data.Accumulatedusage.Storageseconds) + storageCosts(*data.Accumulatedusage.Storageseconds)
	}
	var lifetime string
	if data.Accumulatedusage.Lifetime != nil {
		lifetime = humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime))
	}
	footer := []string{"Total",
		objects,
		storage,
		lifetime,
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a container usage as table
func (s ContainerBillingTablePrinter) Print(data *models.V1ContainerUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "Namespace", "PodUUID", "PodName", "PodStartDate", "PodEndDate", "ContainerName", "ContainerImage", "Lifetime", "CPUSeconds", "MemorySeconds"}
	s.shortHeader = []string{"Tenant", "ProjectID", "Partition", "ClusterName", "Namespace", "PodName", "ContainerName", "Lifetime", "CPU (1 * s)", "Memory (Gi * h)"}
	if s.order == "" {
		s.order = "tenant,project,partition,cluster,namespace,pod,container"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var namespace string
		if u.Namespace != nil {
			namespace = *u.Namespace
		}
		var podUUID string
		if u.Poduuid != nil {
			podUUID = *u.Poduuid
		}
		var podName string
		if u.Podname != nil {
			podName = *u.Podname
		}
		var podStart string
		if u.Podstart != nil {
			podStart = u.Podstart.String()
		}
		var podEnd string
		if u.Podend != nil {
			podEnd = u.Podend.String()
		}
		var containerName string
		if u.Containername != nil {
			containerName = *u.Containername
		}
		var containerImage string
		if u.Containerimage != nil {
			containerImage = *u.Containerimage
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		var cpuUsage string
		if u.Cpuseconds != nil {
			cpuUsage = humanizeCPU(*u.Cpuseconds)
		}
		var memoryUsage string
		if u.Memoryseconds != nil {
			memoryUsage = humanizeMemory(*u.Memoryseconds)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			clusterID,
			clusterName,
			namespace,
			podUUID,
			podName,
			podStart,
			podEnd,
			containerName,
			containerImage,
			humanizeDuration(lifetime),
			cpuUsage,
			memoryUsage,
		}
		short := []string{
			tenant,
			projectID,
			partition,
			clusterName,
			namespace,
			podName,
			containerName,
			humanizeDuration(lifetime),
			cpuUsage,
			memoryUsage,
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)),
		humanizeCPU(*data.Accumulatedusage.Cpuseconds) + cpuCosts(*data.Accumulatedusage.Cpuseconds),
		humanizeMemory(*data.Accumulatedusage.Memoryseconds) + memoryCosts(*data.Accumulatedusage.Memoryseconds),
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

// Print a postgres usage as table
func (s PostgresBillingTablePrinter) Print(data *models.V1PostgresUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "PostgresID", "Description", "Start", "End", "CPU (1 * s)", "Memory (Gi * h)", "StorageSeconds (Gi * h)", "Lifetime"}
	s.shortHeader = []string{"Tenant", "ProjectID", "PostgresID", "Description", "CPU (1 * s)", "Memory (Gi * h)", "StorageSeconds (Gi * h)", "Lifetime"}
	if s.order == "" {
		s.order = "tenant,project,id"
	}
	s.Order(data.Usage)
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var postgresID string
		if u.Postgresid != nil {
			postgresID = *u.Postgresid
		}
		var postgresDescription string
		if u.Postgresdescription != nil {
			postgresDescription = *u.Postgresdescription
		}
		var start string
		if u.Postgresstart != nil {
			start = u.Postgresstart.String()
		}
		var end string
		if u.Postgresend != nil {
			end = u.Postgresend.String()
		}
		var cpu string
		if u.Cpuseconds != nil {
			cpu = humanizeCPU(*u.Cpuseconds)
		}
		var memory string
		if u.Memoryseconds != nil {
			memory = humanizeMemory(*u.Memoryseconds)
		}
		var storage string
		if u.Storageseconds != nil {
			storage = humanizeMemory(*u.Storageseconds)
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			postgresID,
			postgresDescription,
			start,
			end,
			cpu,
			memory,
			storage,
			humanizeDuration(lifetime),
		}
		short := []string{
			tenant,
			projectID,
			postgresID,
			postgresDescription,
			cpu,
			memory,
			storage,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeCPU(*data.Accumulatedusage.Cpuseconds) + cpuCosts(*data.Accumulatedusage.Cpuseconds),
		humanizeMemory(*data.Accumulatedusage.Memoryseconds) + memoryCosts(*data.Accumulatedusage.Memoryseconds),
		humanizeMemory(*data.Accumulatedusage.Storageseconds) + storageCosts(*data.Accumulatedusage.Storageseconds),
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)),
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)   // nolint:makezero
	s.addShortData(append(shortFooter, footer...), data) // nolint:makezero
	s.render()
}

func humanizeBytesToGi(amountInBytes string) string {
	i := new(big.Float)
	i.SetString(amountInBytes)
	gi := new(big.Float).Quo(i, big.NewFloat(1<<30))
	return fmt.Sprintf("%.2f", gi)
}

func humanizeMemory(memorySeconds string) string {
	i := new(big.Float)
	i.SetString(memorySeconds)
	ms := new(big.Float).Quo(i, big.NewFloat(1<<30))
	memoryHours := new(big.Float).Quo(ms, big.NewFloat(3600))
	return fmt.Sprintf("%.2f", memoryHours)
}

func humanizeCPU(cpuSeconds string) string {
	duration, err := strconv.ParseInt(cpuSeconds, 10, 64)
	if err == nil {
		return humanizeDuration(time.Duration(duration) * time.Second)
	}
	return ""
}

func lifetimeCosts(lifetime int64) string {
	perHour := viper.GetFloat64("costs-hour")
	if perHour <= 0 {
		return ""
	}
	return fmt.Sprintf(" (%.2f €)", time.Duration(lifetime).Hours()*perHour)
}

func cpuCosts(cpuSeconds string) string {
	cpuPerCoreAndHour := viper.GetFloat64("costs-cpu-hour")
	if cpuPerCoreAndHour <= 0 {
		return ""
	}
	duration, err := strconv.ParseInt(cpuSeconds, 10, 64)
	if err == nil {
		return fmt.Sprintf(" (%.2f €)", float64(duration/3600)*cpuPerCoreAndHour)
	}
	return ""
}

func memoryCosts(memorySeconds string) string {
	memoryPerGiAndHour := viper.GetFloat64("costs-memory-gi-hour")
	if memoryPerGiAndHour <= 0 {
		return ""
	}
	i := new(big.Float)
	i.SetString(memorySeconds)
	ms := new(big.Float).Quo(i, big.NewFloat(1<<30))
	memoryHours := new(big.Float).Quo(ms, big.NewFloat(3600))
	memoryCosts := new(big.Float).Mul(memoryHours, big.NewFloat(memoryPerGiAndHour))
	return fmt.Sprintf(" (%.2f €)", memoryCosts)
}

func giCosts(amountInBytes string, costsPerGi float64) string {
	if costsPerGi <= 0 {
		return ""
	}

	i := new(big.Float)
	i.SetString(amountInBytes)
	gi := new(big.Float).Quo(i, big.NewFloat(1<<30))
	costs := new(big.Float).Mul(gi, big.NewFloat(costsPerGi))
	return fmt.Sprintf(" (%.2f €)", costs)
}

func storageCosts(storageSeconds string) string {
	storagePerGiAndHour := viper.GetFloat64("costs-storage-gi-hour")
	if storagePerGiAndHour <= 0 {
		return ""
	}
	i := new(big.Float)
	i.SetString(storageSeconds)
	ss := new(big.Float).Quo(i, big.NewFloat(1<<30))
	storageHours := new(big.Float).Quo(ss, big.NewFloat(3600))
	storageCosts := new(big.Float).Mul(storageHours, big.NewFloat(storagePerGiAndHour))
	return fmt.Sprintf(" (%.2f €)", storageCosts)
}
