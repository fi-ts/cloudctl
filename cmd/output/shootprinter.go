package output

import (
	"fmt"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/spf13/viper"
)

type (
	// ShootTablePrinter print a Shoot Cluster in a Table
	ShootTablePrinter struct {
		TablePrinter
	}
	ShootTableDetailPrinter struct {
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
	ImageExpirationDaysDefault = 7
)

type shootStats struct {
	apiServer    string
	controlPlane string
	nodes        string
	system       string
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

func imageExpires(id string, expirationDate string) error {
	t, err := time.Parse(time.RFC3339, expirationDate)
	if err != nil {
		return fmt.Errorf("Image has no expiration date set: %s", id)
	}

	viper.SetDefault("image-expiration-warning-days", ImageExpirationDaysDefault)
	expirationWarningDays := viper.GetInt("image-expiration-warning-days")
	expiresInHours := int(time.Until(t).Hours())
	if expiresInHours > 0 && expiresInHours < expirationWarningDays*24 {
		return fmt.Errorf("Image expires in %d day(s): %s", expiresInHours/24, id)
	} else if expiresInHours < 0 {
		return fmt.Errorf("Image has expired since %d day(s): %s", -expiresInHours/24, id)
	}
	return nil
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
	s.wideHeader = []string{"UID", "", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Runtime", "Firewall"}
	s.shortHeader = []string{"UID", "", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

	for _, shoot := range data {
		short, wide, _ := shootData(shoot)
		s.addWideData(wide, shoot)
		s.addShortData(short, shoot)
	}
	s.render()

	if len(data) == 1 {
		fmt.Println("\nRequired Actions:")
		//YAMLPrinter{}.Print(actions)
	}
}

// Print a Shoot as table
func (s ShootTableDetailPrinter) Print(shoot *models.V1ClusterResponse) {
	s.wideHeader = []string{"UID", "", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Runtime", "Firewall"}
	s.shortHeader = []string{"UID", "", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}

	short, wide, actions := shootData(shoot)
	s.addWideData(wide, shoot)
	s.addShortData(short, shoot)

	s.render()

	fmt.Println("\nRequired Actions:")
	YAMLPrinter{}.Print(actions)
}

func shootData(shoot *models.V1ClusterResponse) ([]string, []string, []string) {
	shootStats := newShootStats(shoot.Status)

	maintainEmoji := ""
	var actions []string
	for _, m := range shoot.Machines {
		if m.Allocation != nil && m.Allocation.Image != nil && m.Allocation.Image.ExpirationDate != nil {
			expires := imageExpires(*m.Allocation.Image.ID, *m.Allocation.Image.ExpirationDate)
			if expires != nil {
				actions = append(actions, expires.Error())
			}
		}
		// TODO: Check Kubernetes version expiration
		// TODO: Add check for MCM OOT migration
	}
	if len(actions) > 0 {
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

	runtime := "docker"
	autoScaleMin := int32(0)
	autoScaleMax := int32(0)
	if shoot.Workers != nil && len(shoot.Workers) > 0 {
		workers := shoot.Workers[0]
		autoScaleMin = *workers.Minimum
		autoScaleMax = *workers.Maximum
		if workers.CRI != nil && *workers.CRI != "" {
			runtime = *workers.CRI
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
		maintainEmoji,
		*shoot.Name,
		version, partition, dnsdomain,
		operation,
		progress,
		shootStats.apiServer, shootStats.controlPlane, shootStats.nodes, shootStats.system,
		size,
		age,
		purpose,
		privileged,
		runtime,
		firewallImage,
	}
	short := []string{
		*shoot.ID,
		maintainEmoji,
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

	return short, wide, actions
}
