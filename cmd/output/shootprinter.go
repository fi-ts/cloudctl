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
		TablePrinter
	}
	ShootIssuesTablePrinter struct {
		TablePrinter
	}
	// ShootConditionsTablePrinter print the Conditions of a Shoot Cluster in a Table
	ShootConditionsTablePrinter struct {
		TablePrinter
	}

	ShootLastErrorsTablePrinter struct {
		TablePrinter
	}

	ShootLastOperationTablePrinter struct {
		TablePrinter
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
	s.wideHeader = []string{"UID", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Runtime", "Firewall"}
	s.shortHeader = []string{"UID", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

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
	s.wideHeader = []string{"UID", "", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Runtime", "Firewall"}
	s.shortHeader = []string{"UID", "", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

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
	mcmMigrated := false
	for _, feature := range shoot.ControlPlaneFeatureGates {
		if feature == "machineControllerManagerOOT" {
			mcmMigrated = true
			break
		}
	}
	if !mcmMigrated {
		issues = append(issues, "Cluster requires migration to out-of-tree machine-controller-manager, please enable via shoot spec")
	}

	if len(issues) > 0 {
		maintainEmoji = "⚠️"
	}

	age := ""
	if shoot.CreationTimestamp != nil {
		age = helper.HumanizeDuration(time.Since(time.Time(*shoot.CreationTimestamp)))
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

	runtimes := []string{"docker"}
	autoScaleMin := int32(0)
	autoScaleMax := int32(0)
	for _, w := range shoot.Workers {
		autoScaleMin += *w.Minimum
		autoScaleMax += *w.Maximum
		if w.CRI != nil && *w.CRI != "" {
			runtimes = append(runtimes, *w.CRI)
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
	wide := []string{
		*shoot.ID,
		*shoot.Name,
		version, partition, dnsdomain,
		operation,
		progress,
		shootStats.apiServer, shootStats.controlPlane, shootStats.nodes, shootStats.system,
		size,
		age,
		purpose,
		privileged,
		strings.Join(uniqueStringSlice(runtimes), "\n"),
		firewallImage,
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
		return fmt.Errorf("Image of %q has no valid expiration date: %s", host, imageID)
	}

	if t.IsZero() {
		return nil
	}

	viper.SetDefault("image-expiration-warning-days", ImageExpirationDaysDefault)
	expirationWarningDays := viper.GetInt("image-expiration-warning-days")
	expiresInHours := int(time.Until(t).Hours())

	if expiresInHours <= 0 {
		return fmt.Errorf("Image of %q has expired since %d day(s): %s", host, -expiresInHours/24, imageID)
	} else if expiresInHours < expirationWarningDays*24 {
		return fmt.Errorf("Image of %q expires in %d day(s): %s", host, expiresInHours/24, imageID)
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
		return fmt.Errorf("Kubernetes support has expired since %d day(s): %s", -expiresInHours/24, *shoot.Kubernetes.Version)
	} else if expiresInHours < expirationWarningDays*24 {
		return fmt.Errorf("Kubernetes support expires in %d day(s): %s", expiresInHours/24, *shoot.Kubernetes.Version)
	}

	return nil
}
