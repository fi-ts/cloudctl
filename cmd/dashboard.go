package cmd

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/health"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"
	"k8s.io/utils/pointer"
)

var (
	dashboardCmd = &cobra.Command{
		Use:   "dashboard",
		Short: "shows a live dashboard optimized for operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard()
		},
		PreRun: bindPFlags,
	}

	dashboardErr error
)

func init() {
	dashboardCmd.Flags().String("partition", "", "show clusters in partition [optional]")
	dashboardCmd.Flags().String("tenant", "", "show clusters of given tenant [optional]")
	dashboardCmd.Flags().String("purpose", "", "show clusters of given purpose [optional]")
	dashboardCmd.Flags().String("color-theme", "default", "the dashboard's color theme [default|dark] [optional]")
	dashboardCmd.Flags().Duration("refresh-interval", 3*time.Second, "refresh interval [optional]")

	err := dashboardCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("tenant", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tenantListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("purpose", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"production", "development", "evaluation"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("color-theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"default\twith bright fonts, optimized for dark terminal backgrounds",
			"dark\twith dark fonts, optimized for bright terminal backgrounds",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func dashboardApplyTheme(theme string) error {
	switch theme {
	case "default":
		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorWhite)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	case "dark":
		ui.Theme.Default = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Border = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Title = ui.NewStyle(ui.ColorBlack)

		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorBlack)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorBlack)}

		ui.Theme.Gauge.Label = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Gauge.Label.Fg = ui.ColorBlack

		ui.Theme.Paragraph.Text = ui.NewStyle(ui.ColorBlack)

		ui.Theme.Table.Text = ui.NewStyle(ui.ColorBlack)
	default:
		return fmt.Errorf("unknown theme: %s", theme)
	}
	return nil
}

func runDashboard() error {
	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	err := dashboardApplyTheme(viper.GetString("color-theme"))
	if err != nil {
		return err
	}

	var (
		interval      = viper.GetDuration("refresh-interval")
		width, height = ui.TerminalDimensions()
		d             = NewDashboard()
	)

	d.Size(0, 0, width, height)
	d.Render()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(interval)

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "1":
				d.tabPane.FocusLeft()
				ui.Clear()
				d.Render()
			case "2":
				d.tabPane.FocusRight()
				ui.Clear()
				d.Render()
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				var (
					height = payload.Height
					width  = payload.Width
				)
				d.Size(0, 0, width, height)
				ui.Clear()
				d.Render()
			}
		case <-ticker.C:
			d.Render()
		}
	}
}

type dashboard struct {
	statusHeader *widgets.Paragraph
	filterHeader *widgets.Paragraph

	tabPane *widgets.TabPane

	clusterPane *dashboardClusterPane
	volumePane  *dashboardVolumePane

	sem *semaphore.Weighted
}

func NewDashboard() *dashboard {
	var (
		tenant    = viper.GetString("tenant")
		partition = viper.GetString("partition")
		purpose   = viper.GetString("purpose")
	)

	d := &dashboard{}

	d.sem = semaphore.NewWeighted(1)

	d.clusterPane = NewDashboardClusterPane()
	d.volumePane = NewDashboardVolumePane()

	d.tabPane = widgets.NewTabPane("(1) Clusters", "(2) Volumes")
	d.tabPane.ActiveTabStyle = ui.NewStyle(ui.ColorBlue)
	d.tabPane.Border = false

	d.statusHeader = widgets.NewParagraph()
	d.statusHeader.Title = "Cloud Dashboard"
	d.statusHeader.WrapText = false

	d.filterHeader = widgets.NewParagraph()
	d.filterHeader.Title = "Filters"
	d.filterHeader.Text = fmt.Sprintf("Tenant=%s\nPartition=%s\nPurpose=%s", tenant, partition, purpose)
	d.filterHeader.WrapText = false

	return d
}

func (d *dashboard) Size(x1, y1, x2, y2 int) {
	d.statusHeader.SetRect(x1, 0, x2-25, d.headerHeight()-1)
	d.filterHeader.SetRect(x2-25, 0, x2, d.headerHeight()-1)
	d.tabPane.SetRect(x1, d.headerHeight()-1, x2, d.headerHeight())

	d.clusterPane.Size(0, d.headerHeight(), x2, y2)
	d.volumePane.Size(0, d.headerHeight(), x2, y2)
}

func (d *dashboard) headerHeight() int {
	return 6
}

func (d *dashboard) Render() {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return
	}
	defer d.sem.Release(1)

	ui.Render(d.filterHeader, d.tabPane)

	var (
		apiVersion       = "unknown"
		apiHealth        = "unknown"
		apiHealthMessage string
	)

	defer func() {
		var coloredHealth string
		switch apiHealth {
		case rest.HealthStatusHealthy:
			coloredHealth = "[" + apiHealth + "](fg:green)"
		case rest.HealthStatusUnhealthy:
			if apiHealthMessage != "" {
				coloredHealth = "[" + apiHealth + fmt.Sprintf(" (%s)](fg:red)", apiHealthMessage)
			} else {
				coloredHealth = "[" + apiHealth + "](fg:red)"
			}
		default:
			coloredHealth = apiHealth
		}

		versionLine := fmt.Sprintf("cloud-api %s (API Health: %s), cloudctl %s (%s)", apiVersion, coloredHealth, v.Version, v.GitSHA1)
		fetchInfoLine := fmt.Sprintf("Last Update: %s", time.Now().Format("15:04:05"))
		if dashboardErr != nil {
			fetchInfoLine += fmt.Sprintf(", [Update Error: %s](fg:red)", dashboardErr)
		}
		glossaryLine := "Switch between tabs with number keys. Press q to quit."

		d.statusHeader.Text = fmt.Sprintf("%s\n%s\n%s", versionLine, fetchInfoLine, glossaryLine)
		ui.Render(d.statusHeader)
	}()

	var infoResp *version.InfoOK
	infoResp, dashboardErr = cloud.Version.Info(version.NewInfoParams(), nil)
	if dashboardErr != nil {
		return
	}
	apiVersion = *infoResp.Payload.Version

	var healthResp *health.HealthOK
	healthResp, dashboardErr = cloud.Health.Health(health.NewHealthParams(), nil)
	if dashboardErr != nil {
		return
	}
	apiHealth = *healthResp.Payload.Status
	apiHealthMessage = *healthResp.Payload.Message

	switch d.tabPane.ActiveTabIndex {
	case 0:
		d.clusterPane.Render()
	case 1:
		d.volumePane.Render()
	}
}

type dashboardClusterError struct {
	ClusterName  string
	ErrorMessage string
	Time         time.Time
}

type dashboardClusterPane struct {
	clusterHealth *widgets.BarChart

	clusterStatusAPI     *widgets.Gauge
	clusterStatusControl *widgets.Gauge
	clusterStatusNodes   *widgets.Gauge
	clusterStatusSystem  *widgets.Gauge

	clusterProblems   *widgets.Table
	clusterLastErrors *widgets.Table

	sem *semaphore.Weighted
}

func NewDashboardClusterPane() *dashboardClusterPane {
	d := &dashboardClusterPane{}

	d.sem = semaphore.NewWeighted(1)

	d.clusterHealth = widgets.NewBarChart()
	d.clusterHealth.Labels = []string{"Succeeded", "Progressing", "Unhealthy"}
	d.clusterHealth.Title = "Cluster Operation"
	d.clusterHealth.PaddingLeft = 5
	d.clusterHealth.BarWidth = 5
	d.clusterHealth.BarGap = 10
	d.clusterHealth.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed}

	d.clusterStatusAPI = widgets.NewGauge()
	d.clusterStatusAPI.Title = "API"
	d.clusterStatusAPI.BarColor = ui.ColorGreen

	d.clusterStatusControl = widgets.NewGauge()
	d.clusterStatusControl.Title = "Control"
	d.clusterStatusControl.BarColor = ui.ColorGreen

	d.clusterStatusNodes = widgets.NewGauge()
	d.clusterStatusNodes.Title = "Nodes"
	d.clusterStatusNodes.BarColor = ui.ColorGreen

	d.clusterStatusSystem = widgets.NewGauge()
	d.clusterStatusSystem.Title = "System"
	d.clusterStatusSystem.BarColor = ui.ColorGreen

	d.clusterProblems = widgets.NewTable()
	d.clusterProblems.Title = "Cluster Problems"
	d.clusterProblems.TextAlignment = ui.AlignLeft
	d.clusterProblems.RowSeparator = false

	d.clusterLastErrors = widgets.NewTable()
	d.clusterLastErrors.Title = "Last Errors"
	d.clusterLastErrors.TextAlignment = ui.AlignLeft
	d.clusterLastErrors.RowSeparator = false

	return d
}

func (d *dashboardClusterPane) Size(x1, y1, x2, y2 int) {
	d.clusterHealth.SetRect(x1, y1, x1+48, y1+12)

	d.clusterStatusAPI.SetRect(x1+50, y1, x2, 3+y1)
	d.clusterStatusControl.SetRect(x1+50, 3+y1, x2, 6+y1)
	d.clusterStatusNodes.SetRect(x1+50, 6+y1, x2, 9+y1)
	d.clusterStatusSystem.SetRect(x1+50, 9+y1, x2, 12+y1)

	tableHeights := (y2 - (y1 + 12)) / 2

	d.clusterProblems.SetRect(x1, 12+y1, x2, y1+12+tableHeights)
	d.clusterProblems.ColumnWidths = []int{12, x2 - 12}

	d.clusterLastErrors.SetRect(x1, 12+y1+tableHeights, x2, y2)
	d.clusterLastErrors.ColumnWidths = []int{12, x2 - 12}
}

func (d *dashboardClusterPane) Render() {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return
	}
	defer d.sem.Release(1)

	var (
		tenant    = viper.GetString("tenant")
		partition = viper.GetString("partition")
		purpose   = viper.GetString("purpose")

		clusters    []*models.V1ClusterResponse
		filteredOut int

		succeeded  int
		processing int
		unhealthy  int

		apiOK     int
		controlOK int
		nodesOK   int
		systemOK  int

		clusterErrors []dashboardClusterError
		lastErrors    []dashboardClusterError

		strDeref = func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}
	)

	var resp *cluster.FindClustersOK
	resp, dashboardErr = cloud.Cluster.FindClusters(cluster.NewFindClustersParams().WithBody(&models.V1ClusterFindRequest{
		PartitionID: strDeref(partition),
		Tenant:      strDeref(tenant),
	}).WithReturnMachines(pointer.BoolPtr(false)), nil)
	if dashboardErr != nil {
		return
	}
	clusters = resp.Payload

	for _, c := range clusters {
		if c.Purpose == nil || (purpose != "" && *c.Purpose != purpose) {
			filteredOut++
			continue
		}
		if c.Status == nil || c.Status.LastOperation == nil || c.Status.LastOperation.State == nil || *c.Status.LastOperation.State == "" {
			unhealthy++
			continue
		}

		switch *c.Status.LastOperation.State {
		case string(v1beta1.LastOperationStateSucceeded):
			succeeded++
		case string(v1beta1.LastOperationStateProcessing):
			processing++
		default:
			unhealthy++
		}

		for _, condition := range c.Status.Conditions {
			if condition == nil || condition.Status == nil || condition.Type == nil {
				continue
			}

			status := *condition.Status
			if status != "True" {
				if c.Name == nil || condition.Message == nil || condition.LastUpdateTime == nil {
					continue
				}
				t, err := time.Parse(time.RFC3339, *condition.LastUpdateTime)
				if err != nil {
					continue
				}
				clusterErrors = append(clusterErrors, dashboardClusterError{
					ClusterName:  *c.Name,
					ErrorMessage: fmt.Sprintf("(%s) %s", *condition.Type, *condition.Message),
					Time:         t,
				})
				continue
			}

			switch *condition.Type {
			case string(v1beta1.ShootControlPlaneHealthy):
				controlOK++
			case string(v1beta1.ShootEveryNodeReady):
				nodesOK++
			case string(v1beta1.ShootSystemComponentsHealthy):
				systemOK++
			case string(v1beta1.ShootAPIServerAvailable):
				apiOK++
			}
		}

		for _, e := range c.Status.LastErrors {
			if c.Name == nil || e.Description == nil {
				continue
			}
			t, err := time.Parse(time.RFC3339, e.LastUpdateTime)
			if err != nil {
				continue
			}
			lastErrors = append(lastErrors, dashboardClusterError{
				ClusterName:  *c.Name,
				ErrorMessage: *e.Description,
				Time:         t,
			})
		}
	}

	processedClusters := len(clusters) - filteredOut
	if processedClusters <= 0 {
		return
	}

	// for some reason the UI hangs when all values are zero...
	// so we render this individually
	if succeeded > 0 || processing > 0 || unhealthy > 0 {
		d.clusterHealth.Data = []float64{float64(succeeded), float64(processing), float64(unhealthy)}
		ui.Render(d.clusterHealth)
	}

	sort.Slice(clusterErrors, func(i, j int) bool {
		return clusterErrors[i].Time.Before(clusterErrors[j].Time)
	})
	rows := [][]string{}
	for _, e := range clusterErrors {
		rows = append(rows, []string{e.ClusterName, e.ErrorMessage})
	}
	d.clusterProblems.Rows = rows
	ui.Render(d.clusterProblems)

	sort.Slice(lastErrors, func(i, j int) bool {
		return lastErrors[i].Time.Before(lastErrors[j].Time)
	})
	rows = [][]string{}
	for _, e := range lastErrors {
		rows = append(rows, []string{e.ClusterName, e.ErrorMessage})
	}
	d.clusterLastErrors.Rows = rows
	ui.Render(d.clusterLastErrors)

	d.clusterStatusAPI.Percent = apiOK * 100 / processedClusters
	d.clusterStatusControl.Percent = controlOK * 100 / processedClusters
	d.clusterStatusNodes.Percent = nodesOK * 100 / processedClusters
	d.clusterStatusSystem.Percent = systemOK * 100 / processedClusters
	ui.Render(d.clusterStatusAPI, d.clusterStatusControl, d.clusterStatusNodes, d.clusterStatusSystem)
}

type dashboardVolumePane struct {
	healthyClusters *widgets.BarChart

	sem *semaphore.Weighted
}

func NewDashboardVolumePane() *dashboardVolumePane {
	d := &dashboardVolumePane{}

	d.sem = semaphore.NewWeighted(1)

	d.healthyClusters = widgets.NewBarChart()
	d.healthyClusters.Labels = []string{"Healthy", "Unhealthy"}
	d.healthyClusters.Title = "Clusters"
	d.healthyClusters.PaddingLeft = 5
	d.healthyClusters.BarWidth = 5
	d.healthyClusters.BarGap = 10
	d.healthyClusters.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed}

	return d
}

func (d *dashboardVolumePane) Size(x1, y1, x2, y2 int) {
	d.healthyClusters.SetRect(x1, y1, x2, y2)
}

func (d *dashboardVolumePane) Render() {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return
	}
	defer d.sem.Release(1)

	var (
		partition = viper.GetString("partition")

		clusters    []*models.V1StorageClusterInfo
		filteredOut int

		healthy   int
		unhealthy int
	)

	var resp *volume.ClusterInfoOK
	resp, dashboardErr = cloud.Volume.ClusterInfo(volume.NewClusterInfoParams(), nil)
	if dashboardErr != nil {
		return
	}
	clusters = resp.Payload

	for _, c := range clusters {
		if c.Partition == nil || (partition != "" && *c.Partition != partition) {
			filteredOut++
			continue
		}
		if c.Health == nil || c.Health.State == nil {
			unhealthy++
			continue
		}

		switch *c.Health.State {
		case "OK":
			healthy++
		default:
			unhealthy++
		}
	}

	processedClusters := len(clusters) - filteredOut
	if processedClusters <= 0 {
		return
	}

	// for some reason the UI hangs when all values are zero...
	// so we render this individually
	if healthy > 0 || unhealthy > 0 {
		d.healthyClusters.Data = []float64{float64(healthy), float64(unhealthy)}
		ui.Render(d.healthyClusters)
	}
}
