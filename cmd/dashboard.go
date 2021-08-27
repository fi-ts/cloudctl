package cmd

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/health"
	"github.com/fi-ts/cloud-go/api/client/version"
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
		tenant        = viper.GetString("tenant")
		partition     = viper.GetString("partition")
		purpose       = viper.GetString("purpose")
		interval      = viper.GetDuration("refresh-interval")
		sem           = semaphore.NewWeighted(1)
		width, height = ui.TerminalDimensions()
		strDeref      = func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}
	)

	type clusterError struct {
		ClusterName  string
		ErrorMessage string
		Time         time.Time
	}

	headerHeight := 5
	header := widgets.NewParagraph()
	header.Title = "Cluster Dashboard"
	header.SetRect(0, 0, width-25, headerHeight)

	filters := widgets.NewParagraph()
	filters.Title = "Filters"
	filters.Text = fmt.Sprintf("Tenant=%s\nPartition=%s\nPurpose=%s", tenant, partition, purpose)
	filters.SetRect(width-25, 0, width, headerHeight)

	clusterHealth := widgets.NewBarChart()
	clusterHealth.Labels = []string{"Succeeded", "Progressing", "Unhealthy"}
	clusterHealth.Title = "Cluster Operation"
	clusterHealth.PaddingLeft = 5
	clusterHealth.SetRect(0, headerHeight, 48, 12+headerHeight)
	clusterHealth.BarWidth = 5
	clusterHealth.BarGap = 10
	clusterHealth.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed}

	clusterStatusAPI := widgets.NewGauge()
	clusterStatusAPI.Title = "API"
	clusterStatusAPI.SetRect(50, headerHeight, width, 3+headerHeight)
	clusterStatusAPI.BarColor = ui.ColorGreen

	clusterStatusControl := widgets.NewGauge()
	clusterStatusControl.Title = "Control"
	clusterStatusControl.SetRect(50, 3+headerHeight, width, 6+headerHeight)
	clusterStatusControl.BarColor = ui.ColorGreen

	clusterStatusNodes := widgets.NewGauge()
	clusterStatusNodes.Title = "Nodes"
	clusterStatusNodes.SetRect(50, 6+headerHeight, width, 9+headerHeight)
	clusterStatusNodes.BarColor = ui.ColorGreen

	clusterStatusSystem := widgets.NewGauge()
	clusterStatusSystem.Title = "System"
	clusterStatusSystem.SetRect(50, 9+headerHeight, width, 12+headerHeight)
	clusterStatusSystem.BarColor = ui.ColorGreen

	tableHeights := (height - (headerHeight + 12)) / 2

	clusterProblems := widgets.NewTable()
	clusterProblems.Title = "Cluster Problems"
	clusterProblems.TextAlignment = ui.AlignLeft
	clusterProblems.RowSeparator = false
	clusterProblems.ColumnWidths = []int{12, width - 12}
	clusterProblems.SetRect(0, 12+headerHeight, width, 12+headerHeight+tableHeights)

	clusterLastErrors := widgets.NewTable()
	clusterLastErrors.Title = "Last Errors"
	clusterLastErrors.TextAlignment = ui.AlignLeft
	clusterLastErrors.RowSeparator = false
	clusterLastErrors.ColumnWidths = []int{12, width - 12}
	clusterLastErrors.SetRect(0, 12+headerHeight+tableHeights, width, height)

	ui.Render(filters)

	refresh := func() {
		if !sem.TryAcquire(1) { // prevent concurrent updates
			return
		}
		defer sem.Release(1)

		var (
			clusters    []*models.V1ClusterResponse
			filteredOut int

			succeeded  int
			processing int
			unhealthy  int

			apiOK     int
			controlOK int
			nodesOK   int
			systemOK  int

			err              error
			apiVersion       = "unknown"
			apiHealth        = "unknown"
			apiHealthMessage string

			clusterErrors []clusterError
			lastErrors    []clusterError
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
			if err != nil {
				fetchInfoLine += fmt.Sprintf(", [Update Error: %s](fg:red)", err)
			}
			glossaryLine := "Press q to quit."

			header.Text = fmt.Sprintf("%s\n%s\n%s", versionLine, fetchInfoLine, glossaryLine)
			ui.Render(header)
		}()

		var infoResp *version.InfoOK
		infoResp, err = cloud.Version.Info(version.NewInfoParams(), nil)
		if err != nil {
			return
		}
		apiVersion = *infoResp.Payload.Version

		var healthResp *health.HealthOK
		healthResp, err = cloud.Health.Health(health.NewHealthParams(), nil)
		if err != nil {
			return
		}
		apiHealth = *healthResp.Payload.Status
		apiHealthMessage = *healthResp.Payload.Message

		var resp *cluster.FindClustersOK
		resp, err = cloud.Cluster.FindClusters(cluster.NewFindClustersParams().WithBody(&models.V1ClusterFindRequest{
			PartitionID: strDeref(partition),
			Tenant:      strDeref(tenant),
		}).WithReturnMachines(pointer.BoolPtr(false)), nil)
		if err != nil {
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
					clusterErrors = append(clusterErrors, clusterError{
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
				lastErrors = append(lastErrors, clusterError{
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
			clusterHealth.Data = []float64{float64(succeeded), float64(processing), float64(unhealthy)}
			ui.Render(clusterHealth)
		}

		sort.Slice(clusterErrors, func(i, j int) bool {
			return clusterErrors[i].Time.Before(clusterErrors[j].Time)
		})
		rows := [][]string{}
		for _, e := range clusterErrors {
			rows = append(rows, []string{e.ClusterName, e.ErrorMessage})
		}
		clusterProblems.Rows = rows
		ui.Render(clusterProblems)

		sort.Slice(lastErrors, func(i, j int) bool {
			return lastErrors[i].Time.Before(lastErrors[j].Time)
		})
		rows = [][]string{}
		for _, e := range lastErrors {
			rows = append(rows, []string{e.ClusterName, e.ErrorMessage})
		}
		clusterLastErrors.Rows = rows
		ui.Render(clusterLastErrors)

		clusterStatusAPI.Percent = apiOK * 100 / processedClusters
		clusterStatusControl.Percent = controlOK * 100 / processedClusters
		clusterStatusNodes.Percent = nodesOK * 100 / processedClusters
		clusterStatusSystem.Percent = systemOK * 100 / processedClusters
		ui.Render(clusterStatusAPI, clusterStatusControl, clusterStatusNodes, clusterStatusSystem)
	}

	refresh()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(interval)

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				width := payload.Width
				height := payload.Height
				clusterStatusAPI.SetRect(55, headerHeight, width, 3+headerHeight)
				clusterStatusControl.SetRect(55, 3+headerHeight, width, 6+headerHeight)
				clusterStatusNodes.SetRect(55, 6+headerHeight, width, 9+headerHeight)
				clusterStatusSystem.SetRect(55, 9+headerHeight, width, 12+headerHeight)
				header.SetRect(0, 0, width-25, headerHeight)
				filters.SetRect(width-25, 0, width, headerHeight)
				tableHeights := (height - (headerHeight + 12)) / 2
				clusterProblems.ColumnWidths = []int{12, width - 12}
				clusterProblems.SetRect(0, 12+headerHeight, width, 12+headerHeight+tableHeights)
				clusterLastErrors.ColumnWidths = []int{12, width - 12}
				clusterLastErrors.SetRect(0, 12+headerHeight+tableHeights, width, height)
				ui.Clear()
				ui.Render(filters)
				refresh()
			}
		case <-ticker.C:
			refresh()
		}
	}
}
