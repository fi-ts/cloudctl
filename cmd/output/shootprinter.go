package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/spf13/viper"
)

type (
	ShootIssuesResponse  *models.V1ClusterResponse
	ShootIssuesResponses []*models.V1ClusterResponse
	// ShootTablePrinter print a Shoot Cluster in a Table
	ShootTablePrinter struct {
		tablePrinter
	}
	ShootIssuesTablePrinter struct {
		tablePrinter
	}
	// ShootConditionsTablePrinter print the Conditions of a Shoot Cluster in a Table
	ShootConditionsTablePrinter struct {
		tablePrinter
	}

	ShootLastErrorsTablePrinter struct {
		tablePrinter
	}

	ShootLastOperationTablePrinter struct {
		tablePrinter
	}
)

const (
	ImageExpirationDaysDefault      = 14
	KuberentesExpirationDaysDefault = 14
)

type shootStats struct {
	apiServer    string
	controlPlane string
	nodes        string
	system       string
}

func (s ShootConditionsTablePrinter) Print(data []*models.V1beta1Condition) {
	s.wideHeader = []string{"LastTransition", "LastUpdate", "Message", "Reason", "Status", "Type"}
	s.shortHeader = []string{"LastTransition", "LastUpdate", "Message", "Reason", "Status", "Type"}
	for _, condition := range data {
		wide := []string{
			strValue(condition.LastTransitionTime),
			strValue(condition.LastUpdateTime),
			strValue(condition.Message),
			strValue(condition.Reason),
			strValue(condition.Status),
			strValue(condition.Type),
		}
		short := wide
		s.addWideData(wide, data)
		s.addShortData(short, data)
	}
	s.render()
}

func (s ShootLastErrorsTablePrinter) Print(data []*models.V1beta1LastError) {
	s.wideHeader = []string{"Time", "Task", "Description"}
	s.shortHeader = []string{"Time", "Task", "Description"}
	for _, e := range data {
		wide := []string{
			strValue(&e.LastUpdateTime),
			strValue(&e.TaskID),
			strValue(e.Description),
		}
		short := wide
		s.addWideData(wide, data)
		s.addShortData(short, data)
	}
	s.render()
}

func (s ShootLastOperationTablePrinter) Print(data *models.V1beta1LastOperation) {
	s.wideHeader = []string{"Time", "State", "Progress", "Description"}
	s.shortHeader = []string{"Time", "State", "Progress", "Description"}
	wide := []string{
		strValue(data.LastUpdateTime),
		strValue(data.State),
		fmt.Sprintf("%d%% [%s]", *data.Progress, *data.Type),
		strValue(data.Description),
	}
	short := wide
	s.addWideData(wide, data)
	s.addShortData(short, data)

	s.render()
}

// Print a Shoot as table
func (s ShootTablePrinter) Print(data []*models.V1ClusterResponse) {
	s.wideHeader = []string{"UID", "Name", "Version", "Partition", "Seed", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "LastUpdate", "Purpose", "Privileged", "Audit", "Runtime", "Firewall", "Firewall Controller", "Log accepted conns", "Egress IPs", "Gardener"}
	s.shortHeader = []string{"UID", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

	if s.order == "" {
		s.order = "tenant,project,name"
	}
	s.Order(data)

	var short []string
	var wide []string
	for _, shoot := range data {
		short, wide, _ = shootData(shoot, false)
		s.addWideData(wide, shoot)
		s.addShortData(short, shoot)
	}
	s.render()
}

func (s ShootIssuesTablePrinter) Print(data []*models.V1ClusterResponse) {
	s.wideHeader = []string{"UID", "", "Name", "Version", "Partition", "Seed", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Audit", "Runtime", "Firewall", "Firewall Controller", "Log accepted conns", "Egress IPs"}
	s.shortHeader = []string{"UID", "", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

	if s.order == "" {
		s.order = "tenant,project,name"
	}
	s.Order(data)

	var short []string
	var wide []string
	var issues []string
	for _, shoot := range data {
		short, wide, issues = shootData(shoot, true)
		s.addWideData(wide, shoot)
		s.addShortData(short, shoot)
	}
	s.render()

	if len(data) == 1 && len(issues) > 0 {
		fmt.Println("\nIssues:")
		printStringSlice(issues)
	}
}

func shootData(shoot *models.V1ClusterResponse, withIssues bool) ([]string, []string, []string) {
	shootStats := newShootStats(shoot.Status)

	maintainEmoji := ""
	var issues []string

	ms := shoot.Machines
	ms = append(ms, shoot.Firewalls...)

	for _, m := range ms {
		expires := imageExpires(m)
		if expires != nil {
			issues = append(issues, expires.Error())
		}
	}

	if shoot.Firewalls != nil {
		switch len(shoot.Firewalls) {
		case 0:
			issues = append(issues, "Cluster has no firewall")
		case 1:
		default:
			issues = append(issues, "Cluster has multiple firewalls, cluster requires manual administration")
		}
	}

	expires := kubernetesExpires(shoot)
	if expires != nil {
		issues = append(issues, expires.Error())
	}

	if len(issues) > 0 {
		maintainEmoji = "⚠️"
	}

	age := ""
	if shoot.CreationTimestamp != nil {
		age = helper.HumanizeDuration(time.Since(time.Time(*shoot.CreationTimestamp)))
	}
	lastReconcilation := ""
	if shoot.Status != nil && shoot.Status.LastOperation != nil && shoot.Status.LastOperation.LastUpdateTime != nil {
		lastUpdate, err := time.Parse(time.RFC3339, *shoot.Status.LastOperation.LastUpdateTime)
		if err != nil {
			lastReconcilation = "unknown"
		} else {
			lastReconcilation = helper.HumanizeDuration(time.Since(lastUpdate))
		}
	}

	gardener := ""
	if shoot.Status != nil && shoot.Status.Gardener != nil && shoot.Status.Gardener.Version != nil {
		gardener = *shoot.Status.Gardener.Version
	}

	operation := ""
	progress := "0%"
	if shoot.Status.LastOperation != nil {
		operation = *shoot.Status.LastOperation.State
		progress = fmt.Sprintf("%d%% [%s]", *shoot.Status.LastOperation.Progress, *shoot.Status.LastOperation.Type)
	}
	partition := ""
	if shoot.PartitionID != nil {
		partition = *shoot.PartitionID
	}
	dnsdomain := ""
	if shoot.DNSEndpoint != nil {
		dnsdomain = *shoot.DNSEndpoint
	}
	version := ""
	if shoot.Kubernetes.Version != nil {
		version = *shoot.Kubernetes.Version
		if shoot.Maintenance != nil && shoot.Maintenance.AutoUpdate != nil && shoot.Maintenance.AutoUpdate.KubernetesVersion != nil && *shoot.Maintenance.AutoUpdate.KubernetesVersion {
			version = fmt.Sprintf("%s↑", version)
		}
	}
	purpose := ""
	if shoot.Purpose != nil {
		p := *shoot.Purpose
		purpose = p[:4]
	}

	privileged := ""
	if shoot.Kubernetes.AllowPrivilegedContainers != nil {
		privileged = fmt.Sprintf("%t", *shoot.Kubernetes.AllowPrivilegedContainers)
	}

	audit := "Off"
	if shoot.ControlPlaneFeatureGates != nil {
		var ca, as bool
		for _, featureGate := range shoot.ControlPlaneFeatureGates {
			switch featureGate {
			case "clusterAudit":
				ca = true
			case "auditToSplunk":
				as = true
			}
		}
		if ca {
			audit = "On"
		}
		if as {
			audit = audit + ",Splunk"
		}
	}

	runtimes := []string{}
	autoScaleMin := int32(0)
	autoScaleMax := int32(0)
	for _, w := range shoot.Workers {
		autoScaleMin += *w.Minimum
		autoScaleMax += *w.Maximum
		if w.CRI != nil && *w.CRI != "" {
			runtimes = append(runtimes, *w.CRI)
		} else {
			runtimes = append(runtimes, "docker")
		}
	}
	currentMachines := "x"
	if shoot.Machines != nil {
		currentMachines = fmt.Sprintf("%d", len(shoot.Machines))
	}
	size := fmt.Sprintf("%d≤%s≤%d", autoScaleMin, currentMachines, autoScaleMax)

	tenant := ""
	if shoot.Tenant != nil {
		tenant = *shoot.Tenant
	}
	project := ""
	if shoot.ProjectID != nil {
		project = *shoot.ProjectID
	}

	firewallImage := ""
	if shoot.FirewallImage != nil {
		firewallImage = *shoot.FirewallImage
	}

	seed := ""
	if shoot.Status != nil {
		seed = shoot.Status.SeedName
	}

	egressIPs := []string{}
	for _, e := range shoot.EgressRules {
		if e == nil {
			continue
		}
		for _, i := range e.IPs {
			egressIPs = append(egressIPs, fmt.Sprintf("%s: %s", *e.NetworkID, i))
		}
	}

	firewallController := ""
	if shoot.FirewallControllerVersion != nil {
		firewallController = *shoot.FirewallControllerVersion
	}

	logAcceptedConnections := ""
	if shoot.ClusterFeatures.LogAcceptedConnections != nil {
		logAcceptedConnections = *shoot.ClusterFeatures.LogAcceptedConnections
	}

	wide := []string{
		*shoot.ID,
		*shoot.Name,
		version, partition, seed, dnsdomain,
		operation,
		progress,
		shootStats.apiServer, shootStats.controlPlane, shootStats.nodes, shootStats.system,
		size,
		age,
		lastReconcilation,
		purpose,
		privileged,
		audit,
		strings.Join(uniqueStringSlice(runtimes), "\n"),
		firewallImage,
		firewallController,
		logAcceptedConnections,
		strings.Join(egressIPs, "\n"),
		gardener,
	}
	short := []string{
		*shoot.ID,
		tenant,
		project,
		*shoot.Name,
		version, partition,
		operation,
		progress,
		shootStats.apiServer, shootStats.controlPlane, shootStats.nodes, shootStats.system,
		size,
		age,
		purpose,
	}

	if withIssues {
		wide = append([]string{*shoot.ID, maintainEmoji}, wide[1:]...)
		short = append([]string{*shoot.ID, maintainEmoji}, short[1:]...)
	}

	return short, wide, issues
}

func newShootStats(status *models.V1beta1ShootStatus) *shootStats {
	res := shootStats{}
	if status != nil {
		for _, condition := range status.Conditions {
			status := *condition.Status
			switch *condition.Type {
			case string(v1beta1.ShootControlPlaneHealthy):
				res.controlPlane = status
			case string(v1beta1.ShootEveryNodeReady):
				res.nodes = status
			case string(v1beta1.ShootSystemComponentsHealthy):
				res.system = status
			case string(v1beta1.ShootAPIServerAvailable):
				res.apiServer = status
			}
		}
	}
	return &res
}

func imageExpires(m *models.ModelsV1MachineResponse) error {
	if m.Allocation == nil || m.Allocation.Image == nil || m.Allocation.Image.ExpirationDate == nil {
		return nil
	}

	host := *m.Allocation.Name
	imageID := *m.Allocation.Image.ID

	t, err := time.Parse(time.RFC3339, *m.Allocation.Image.ExpirationDate)
	if err != nil {
		return fmt.Errorf("image of %q has no valid expiration date: %s", host, imageID)
	}

	if t.IsZero() {
		return nil
	}

	viper.SetDefault("image-expiration-warning-days", ImageExpirationDaysDefault)
	expirationWarningDays := viper.GetInt("image-expiration-warning-days")
	expiresInHours := int(time.Until(t).Hours())

	if expiresInHours <= 0 {
		return fmt.Errorf("image of %q has expired since %d day(s): %s", host, -expiresInHours/24, imageID)
	} else if expiresInHours < expirationWarningDays*24 {
		return fmt.Errorf("image of %q expires in %d day(s): %s", host, expiresInHours/24, imageID)
	}

	return nil
}

func kubernetesExpires(shoot *models.V1ClusterResponse) error {
	if shoot.Kubernetes == nil || shoot.Kubernetes.ExpirationDate == nil || time.Time(*shoot.Kubernetes.ExpirationDate).IsZero() {
		return nil
	}

	viper.SetDefault("kubernetes-expiration-warning-days", ImageExpirationDaysDefault)
	expirationWarningDays := viper.GetInt("kubernetes-expiration-warning-days")
	expiresInHours := int(time.Until(time.Time(*shoot.Kubernetes.ExpirationDate)).Hours())

	if expiresInHours <= 0 {
		return fmt.Errorf("kubernetes support has expired since %d day(s): %s", -expiresInHours/24, *shoot.Kubernetes.Version)
	} else if expiresInHours < expirationWarningDays*24 {
		return fmt.Errorf("kubernetes support expires in %d day(s): %s", expiresInHours/24, *shoot.Kubernetes.Version)
	}

	return nil
}
