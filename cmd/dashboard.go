package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	ui "github.com/dcorbe/termui-dpc"
	"github.com/dcorbe/termui-dpc/widgets"
	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/health"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-lib/pkg/cache"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"

	durosv2 "github.com/metal-stack/duros-go/api/duros/v2"
)

const (
	dashboardRequestsContextTimeout = 5 * time.Second
)

func newDashboardCmd(c *config) *cobra.Command {
	dashboardCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "shows a live dashboard optimized for operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard(c.cloud)
		},
	}

	tabs := dashboardTabs(nil, nil, nil)

	dashboardCmd.Flags().String("partition", "", "show resources in partition [optional]")
	dashboardCmd.Flags().String("tenant", "", "show resources of given tenant [optional]")
	dashboardCmd.Flags().String("purpose", "", "show resources of given purpose [optional]")
	dashboardCmd.Flags().String("color-theme", "default", "the dashboard's color theme [default|dark] [optional]")
	dashboardCmd.Flags().String("initial-tab", strings.ToLower(tabs[0].Name()), "the tab to show when starting the dashboard [optional]")
	dashboardCmd.Flags().Duration("refresh-interval", 3*time.Second, "refresh interval [optional]")

	genericcli.Must(dashboardCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(dashboardCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(dashboardCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))
	genericcli.Must(dashboardCmd.RegisterFlagCompletionFunc("color-theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"default\twith bright fonts, optimized for dark terminal backgrounds",
			"dark\twith dark fonts, optimized for bright terminal backgrounds",
		}, cobra.ShellCompDirectiveNoFileComp
	}))
	genericcli.Must(dashboardCmd.RegisterFlagCompletionFunc("initial-tab", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var names []string
		for _, t := range tabs {
			names = append(names, fmt.Sprintf("%s\t%s", strings.ToLower(t.Name()), t.Description()))
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}))

	return dashboardCmd
}

func dashboardApplyTheme(theme string) error {
	switch theme {
	case "default":
		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorWhite)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorWhite)}

		ui.Theme.Gauge.Bar = ui.ColorWhite

		ui.Theme.Tab.Active = ui.NewStyle(ui.ColorYellow)
	case "dark":
		ui.Theme.Default = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Border = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Title = ui.NewStyle(ui.ColorBlack)

		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorBlack)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorBlack)}

		ui.Theme.Gauge.Label = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Gauge.Label.Fg = ui.ColorBlack
		ui.Theme.Gauge.Bar = ui.ColorBlack

		ui.Theme.Paragraph.Text = ui.NewStyle(ui.ColorBlack)

		ui.Theme.Tab.Active = ui.NewStyle(ui.ColorYellow)
		ui.Theme.Tab.Inactive = ui.NewStyle(ui.ColorBlack)

		ui.Theme.Table.Text = ui.NewStyle(ui.ColorBlack)
	default:
		return fmt.Errorf("unknown theme: %s", theme)
	}
	return nil
}

func runDashboard(cloud *client.CloudAPI) error {
	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	var (
		interval      = viper.GetDuration("refresh-interval")
		width, height = ui.TerminalDimensions()

		uiEvents         = ui.PollEvents()
		eventPassthrough = make(chan ui.Event, 1)
		ticker           = time.NewTicker(interval)
	)

	d, err := NewDashboard(cloud, eventPassthrough)
	if err != nil {
		return err
	}

	d.Resize(0, 0, width, height)
	d.Render()

	panelNumbers := map[string]bool{}
	for i := range d.tabs {
		panelNumbers[strconv.Itoa(i+1)] = true
	}

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				var (
					height = payload.Height
					width  = payload.Width
				)
				d.Resize(0, 0, width, height)
				ui.Clear()
				d.Render()
			default:
				_, ok := panelNumbers[e.ID]
				if ok {
					i, _ := strconv.Atoi(e.ID)
					d.tabPane.ActiveTabIndex = i - 1
					ui.Clear()
					d.Render()
					continue
				}

				// passing on the event to other panels that may wanna use it
				eventPassthrough <- e
			}
		case <-ticker.C:
			d.Render()
		}
	}
}

func dashboardTabs(cache *apiCache, eventChannel chan ui.Event, activePanel func() int) dashboardTabPanes {
	return dashboardTabPanes{
		NewDashboardClusterHealthPane(cache),
		NewDashboardClusterVersionsPane(cache, eventChannel, func() bool {
			return activePanel() == 1
		}),
		NewDashboardVolumePane(cache),
	}
}

type dashboard struct {
	statusHeader *widgets.Paragraph
	filterHeader *widgets.Paragraph

	filterTenant    string
	filterPartition string
	filterPurpose   string

	tabPane *widgets.TabPane
	tabs    dashboardTabPanes

	sem *semaphore.Weighted

	cloud *client.CloudAPI
}

type dashboardTabPane interface {
	Name() string
	Description() string
	Render() error
	Resize(x1, y1, x2, y2 int)
}

type dashboardTabPanes []dashboardTabPane

func (d dashboardTabPanes) FindIndexByName(name string) (int, error) {
	for i, p := range d {
		if strings.EqualFold(p.Name(), name) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("tab with name %q not found", name)
}

func NewDashboard(cloud *client.CloudAPI, eventChannel chan ui.Event) (*dashboard, error) {
	err := dashboardApplyTheme(viper.GetString("color-theme"))
	if err != nil {
		return nil, err
	}

	d := &dashboard{
		sem:             semaphore.NewWeighted(1),
		filterTenant:    viper.GetString("tenant"),
		filterPartition: viper.GetString("partition"),
		filterPurpose:   viper.GetString("purpose"),
		cloud:           cloud,
	}

	d.statusHeader = widgets.NewParagraph()
	d.statusHeader.Title = "Cloud Dashboard"
	d.statusHeader.WrapText = false

	d.filterHeader = widgets.NewParagraph()
	d.filterHeader.Title = "Filters"
	d.filterHeader.WrapText = false

	cache := newCache(cloud,
		viper.GetDuration("refresh-interval"),
		viper.GetString("partition"),
		viper.GetString("tenant"),
		viper.GetString("purpose"),
	)

	d.tabs = dashboardTabs(cache, eventChannel, func() int {
		return d.tabPane.ActiveTabIndex
	})
	var tabNames []string
	for i, p := range d.tabs {
		tabNames = append(tabNames, fmt.Sprintf("(%d) %s", i+1, p.Name()))
	}
	d.tabPane = widgets.NewTabPane(tabNames...)
	d.tabPane.Title = "Tabs"
	d.tabPane.Border = false

	if viper.IsSet("initial-tab") {
		initialPanelIndex, err := d.tabs.FindIndexByName(viper.GetString("initial-tab"))
		if err != nil {
			return nil, err
		}
		d.tabPane.ActiveTabIndex = initialPanelIndex
	}

	return d, nil
}

func (d *dashboard) Resize(x1, y1, x2, y2 int) {
	d.statusHeader.SetRect(x1, y1, x2-25, d.headerHeight())
	d.filterHeader.SetRect(x2-25, y1, x2, d.headerHeight())

	for _, p := range d.tabs {
		p.Resize(x1, d.headerHeight(), x2, y2-1)
	}

	d.tabPane.SetRect(x1, y2-1, x2, y2)
}

func (d *dashboard) headerHeight() int {
	return 5
}

func (d *dashboard) Render() {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return
	}
	defer d.sem.Release(1)

	d.filterHeader.Text = fmt.Sprintf("Tenant=%s\nPartition=%s\nPurpose=%s", d.filterTenant, d.filterPartition, d.filterPurpose)

	ui.Render(d.filterHeader, d.tabPane)

	var (
		apiVersion       = "unknown"
		apiHealth        = "unknown"
		apiHealthMessage string
		releaseVersion   = "unknown"

		lastErr error
	)

	defer func() {
		var coloredHealth string
		switch apiHealth {
		case string(rest.HealthStatusHealthy):
			coloredHealth = "[" + apiHealth + "](fg:green)"
		case string(rest.HealthStatusDegraded), string(rest.HealthStatusPartiallyUnhealthy):
			coloredHealth = "[" + apiHealth + "](fg:yellow)"
		case string(rest.HealthStatusUnhealthy):
			if apiHealthMessage != "" {
				coloredHealth = "[" + apiHealth + fmt.Sprintf(" (%s)](fg:red)", apiHealthMessage)
			} else {
				coloredHealth = "[" + apiHealth + "](fg:red)"
			}
		default:
			coloredHealth = apiHealth
		}

		versionLine := fmt.Sprintf("cloud-api %s (API Health: %s), cloudctl %s (%s), release %s", apiVersion, coloredHealth, v.Version, v.GitSHA1, releaseVersion)
		fetchInfoLine := fmt.Sprintf("Last Update: %s", time.Now().Format("15:04:05"))
		if lastErr != nil {
			fetchInfoLine += fmt.Sprintf(", [Update Error: %s](fg:red)", lastErr)
		}
		glossaryLine := "Switch between tabs with number keys. Press q to quit."

		d.statusHeader.Text = fmt.Sprintf("%s\n%s\n%s", versionLine, fetchInfoLine, glossaryLine)
		ui.Render(d.statusHeader)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), dashboardRequestsContextTimeout)
	defer cancel()

	var infoResp *version.InfoOK
	infoResp, lastErr = d.cloud.Version.Info(version.NewInfoParams().WithContext(ctx), nil)
	if lastErr != nil {
		return
	}
	apiVersion = *infoResp.Payload.Version
	releaseVersion = infoResp.Payload.ReleaseVersion

	healthResp, err := d.cloud.Health.Health(health.NewHealthParams().WithContext(ctx), nil)
	if err != nil {
		var r *health.HealthInternalServerError
		if errors.As(err, &r) {
			healthResp = health.NewHealthOK()
			healthResp.Payload = r.Payload
		} else {
			lastErr = err
			return
		}
	}

	apiHealth = *healthResp.Payload.Status
	apiHealthMessage = *healthResp.Payload.Message

	lastErr = d.tabs[d.tabPane.ActiveTabIndex].Render()
}

type dashboardClusterError struct {
	ClusterName  string
	ErrorMessage string
	Time         time.Time
}

type dashboardClusterHealthPane struct {
	clusterHealth *widgets.BarChart

	clusterStatusAPI     *widgets.Gauge
	clusterStatusControl *widgets.Gauge
	clusterStatusNodes   *widgets.Gauge
	clusterStatusSystem  *widgets.Gauge

	clusterProblems   *widgets.Table
	clusterLastErrors *widgets.Table

	sem *semaphore.Weighted

	cache *apiCache
}

func NewDashboardClusterHealthPane(cache *apiCache) *dashboardClusterHealthPane {
	d := &dashboardClusterHealthPane{}

	d.sem = semaphore.NewWeighted(1)

	d.clusterHealth = widgets.NewBarChart()
	d.clusterHealth.Labels = []string{"Succeeded", "Progressing", "Unhealthy"}
	d.clusterHealth.Title = "Cluster Operation"
	d.clusterHealth.PaddingLeft = 5
	d.clusterHealth.BarWidth = 5
	d.clusterHealth.BarGap = 10
	d.clusterHealth.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed, ui.ColorBlue}
	if viper.GetString("color-theme") == "default" {
		d.clusterHealth.NumStyles = []ui.Style{{Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}, {Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}}
	}

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

	d.cache = cache
	return d
}

func (d *dashboardClusterHealthPane) Name() string {
	return "Clusters"
}

func (d *dashboardClusterHealthPane) Description() string {
	return "Cluster health and issues"
}

func (d *dashboardClusterHealthPane) Resize(x1, y1, x2, y2 int) {
	d.clusterHealth.SetRect(x1, y1, x1+50, y1+12)

	d.clusterStatusAPI.SetRect(x1+50, y1, x2, 3+y1)
	d.clusterStatusControl.SetRect(x1+50, 3+y1, x2, 6+y1)
	d.clusterStatusNodes.SetRect(x1+50, 6+y1, x2, 9+y1)
	d.clusterStatusSystem.SetRect(x1+50, 9+y1, x2, 12+y1)

	tableHeights := int(math.Ceil((float64(y2) - (float64(y1) + 12)) / 2))

	d.clusterProblems.SetRect(x1, 12+y1, x2, y1+12+tableHeights)
	d.clusterProblems.ColumnWidths = []int{12, x2 - 12}

	d.clusterLastErrors.SetRect(x1, 12+y1+tableHeights, x2, y2)
	d.clusterLastErrors.ColumnWidths = []int{12, x2 - 12}
}

func (d *dashboardClusterHealthPane) Render() error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
		clusters []*models.V1ClusterResponse

		succeeded  int
		processing int
		unhealthy  int

		apiOK     int
		controlOK int
		nodesOK   int
		systemOK  int

		clusterErrors []dashboardClusterError
		lastErrors    []dashboardClusterError
	)

	ctx, cancel := context.WithTimeout(context.Background(), dashboardRequestsContextTimeout)
	defer cancel()

	clusters, err := d.cache.Clusters(ctx)
	if err != nil {
		return err
	}

	for _, c := range clusters {
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
			if status != string(v1beta1.ConditionTrue) && status != string(v1beta1.ConditionProgressing) {
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

	processedClusters := len(clusters)
	if processedClusters <= 0 {
		return nil
	}

	// for some reason the UI hangs when all values are zero...
	if succeeded > 0 || processing > 0 || unhealthy > 0 {
		d.clusterHealth.Data = []float64{float64(succeeded), float64(processing), float64(unhealthy)}
		ui.Render(d.clusterHealth)
	}

	sort.Slice(clusterErrors, func(i, j int) bool {
		return clusterErrors[i].Time.After(clusterErrors[j].Time)
	})
	rows := [][]string{}
	for _, e := range clusterErrors {
		rows = append(rows, []string{e.ClusterName, e.ErrorMessage})
	}
	d.clusterProblems.Rows = rows
	ui.Render(d.clusterProblems)

	sort.Slice(lastErrors, func(i, j int) bool {
		return lastErrors[i].Time.After(lastErrors[j].Time)
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

	return nil
}

type dashboardClusterVersionsPane struct {
	kubernetesVersions         *widgets.Tree
	reconciledGardenerVersions *widgets.PieChart
	firewallControllerVersions *widgets.PieChart

	kubernetesVersionsState *kubernetesVersions

	sem *semaphore.Weighted

	cache *apiCache
}

func NewDashboardClusterVersionsPane(cache *apiCache, eventChannel chan ui.Event, isActive func() bool) *dashboardClusterVersionsPane {
	d := &dashboardClusterVersionsPane{}

	d.sem = semaphore.NewWeighted(1)

	d.reconciledGardenerVersions = widgets.NewPieChart()
	d.reconciledGardenerVersions.Title = "Reconciled Gardener Versions"
	d.reconciledGardenerVersions.PaddingTop = 1
	d.reconciledGardenerVersions.PaddingBottom = 1
	d.reconciledGardenerVersions.PaddingRight = 5
	d.reconciledGardenerVersions.PaddingRight = 5

	d.firewallControllerVersions = widgets.NewPieChart()
	d.firewallControllerVersions.Title = "Firewall-Controller Versions"
	d.firewallControllerVersions.PaddingTop = 1
	d.firewallControllerVersions.PaddingBottom = 1
	d.firewallControllerVersions.PaddingLeft = 5
	d.firewallControllerVersions.PaddingRight = 5

	d.kubernetesVersions = widgets.NewTree()
	d.kubernetesVersions.Title = "Kubernetes Versions"
	d.kubernetesVersions.TextStyle = ui.NewStyle(ui.ColorYellow)
	d.kubernetesVersionsState = &kubernetesVersions{}

	go func() {
		for {
			e := <-eventChannel

			if isActive != nil && !isActive() {
				continue
			}
			if !d.kubernetesVersionsState.runOnce {
				continue
			}

			switch e.ID {
			case "<Down>":
				d.kubernetesVersions.ScrollDown()
			case "<Up>":
				d.kubernetesVersions.ScrollUp()
			case "<Space>", "<Enter>":
				d.kubernetesVersions.ToggleExpand()
			case "<PageUp>":
				d.kubernetesVersions.ScrollTop()
			case "<PageDown>":
				d.kubernetesVersions.ScrollBottom()
			case "<Right>":
				d.kubernetesVersions.Expand()
			case "<Left>":
				d.kubernetesVersions.Collapse()
			case "E":
				d.kubernetesVersions.ExpandAll()
			case "C":
				d.kubernetesVersions.CollapseAll()
			}

			ui.Render(d.kubernetesVersions)
		}
	}()

	d.cache = cache

	return d
}

func (d *dashboardClusterVersionsPane) Name() string {
	return "ClusterVersions"
}

func (d *dashboardClusterVersionsPane) Description() string {
	return "Cluster version information"
}

func (d *dashboardClusterVersionsPane) Resize(x1, y1, x2, y2 int) {
	d.kubernetesVersions.SetRect(x1, y1, x2-40, y2)

	tableHeights := int(math.Ceil((float64(y2) - (float64(y1))) / 2))

	d.reconciledGardenerVersions.SetRect(x2-40, y1, x2, y1+tableHeights)
	d.firewallControllerVersions.SetRect(x2-40, y1+tableHeights, x2, y2)
}

type pieVersion map[string]int

func (p pieVersion) total() int {
	var result int
	for _, amount := range p {
		result += amount
	}
	return result
}

func (p pieVersion) toValues() ([]float64, widgets.PieChartLabel) {
	type version struct {
		parsedVersion *semver.Version
		rawVersion    string
		percentage    float64
	}

	var versions []version

	for rawVersion, count := range p {
		v, _ := semver.NewVersion(rawVersion)

		versions = append(versions, version{
			parsedVersion: v,
			rawVersion:    rawVersion,
			percentage:    float64(count) / float64(p.total()),
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		if versions[i].parsedVersion == nil && versions[j].parsedVersion == nil {
			return versions[i].rawVersion < versions[j].rawVersion
		}

		if versions[i].parsedVersion == nil {
			return false
		}
		if versions[j].parsedVersion == nil {
			return true
		}

		return versions[i].parsedVersion.LessThan(versions[j].parsedVersion)
	})

	var result []float64

	for _, v := range versions {
		result = append(result, v.percentage)
	}

	return result, func(dataIndex int, _ float64) string {
		d := versions[dataIndex]
		return fmt.Sprintf("%s (%d%%)", d.rawVersion, int(100*d.percentage))
	}
}

type nodeValue string

func (v nodeValue) String() string {
	return string(v)
}

type kubernetesVersions struct {
	runOnce  bool
	previous []*widgets.TreeNode
}

func (k *kubernetesVersions) update(clusters []*models.V1ClusterResponse) ([]*widgets.TreeNode, bool) {
	var (
		result []*widgets.TreeNode

		sortNodesFunc = func(nodes []*widgets.TreeNode) func(i, j int) bool {
			return func(i, j int) bool {
				iParsed, _ := semver.NewVersion(nodes[i].Value.String())
				jParsed, _ := semver.NewVersion(nodes[j].Value.String())

				if iParsed == nil && jParsed == nil {
					return nodes[i].Value.String() < nodes[j].Value.String()
				}

				if iParsed == nil {
					return false
				}
				if jParsed == nil {
					return true
				}

				return iParsed.LessThan(jParsed)
			}
		}

		findNode = func(nodes []*widgets.TreeNode, name string) *widgets.TreeNode {
			var (
				tree  = &widgets.Tree{}
				found *widgets.TreeNode
			)
			tree.SetNodes(nodes)
			tree.Walk(func(n *widgets.TreeNode) bool {
				if n.Value.String() == name {
					found = n
					return false
				}
				return true
			})
			return found
		}

		ensureNode = func(name string) *widgets.TreeNode {
			if existing := findNode(result, name); existing != nil {
				return existing
			}

			if previous := findNode(k.previous, name); previous != nil {
				result = append(result, previous)
				return previous
			}

			newNode := &widgets.TreeNode{
				Value: nodeValue(name),
			}

			result = append(result, newNode)

			sort.Slice(result, sortNodesFunc(result))

			return newNode
		}

		ensureChildNode = func(parent *widgets.TreeNode, childName string) *widgets.TreeNode {
			if existing := findNode(parent.Nodes, childName); existing != nil {
				return existing
			}

			newNode := &widgets.TreeNode{
				Value: nodeValue(childName),
			}

			parent.Nodes = append(parent.Nodes, newNode)

			sort.Slice(parent.Nodes, sortNodesFunc(parent.Nodes))

			return newNode
		}
	)

	defer func() {
		k.previous = result
	}()

	majorMinorAmounts := map[string]int{}
	versionAmounts := map[string]int{}

	for _, c := range clusters {
		c := c

		var (
			version = pointer.SafeDeref(pointer.SafeDeref(c.Kubernetes).Version)
		)

		majorMinor := "unknown"
		v, _ := semver.NewVersion(version)
		if v != nil {
			majorMinor = fmt.Sprintf("%d.%d", v.Major(), v.Minor())
		}

		majorMinorAmounts[majorMinor]++
		versionAmounts[version]++
	}

	for _, c := range clusters {
		c := c

		var (
			version = pointer.SafeDeref(pointer.SafeDeref(c.Kubernetes).Version)
			tenant  = pointer.SafeDeref(c.Tenant)
			project = pointer.SafeDeref(c.ProjectID)
			id      = pointer.SafeDeref(c.ID)
			name    = pointer.SafeDeref(c.Name)
		)
		if version == "" || tenant == "" || project == "" || id == "" || name == "" {
			continue
		}

		majorMinor := "unknown"
		v, _ := semver.NewVersion(version)
		if v != nil {
			majorMinor = fmt.Sprintf("%d.%d", v.Major(), v.Minor())
		}

		majorMinorNode := ensureNode(fmt.Sprintf("%s (%d)", majorMinor, majorMinorAmounts[majorMinor]))
		versionNode := ensureChildNode(majorMinorNode, fmt.Sprintf("%s (%d)", version, versionAmounts[version]))
		tenantNode := ensureChildNode(versionNode, tenant)
		projectNode := ensureChildNode(tenantNode, project)
		_ = ensureChildNode(projectNode, fmt.Sprintf("%s (%s)", name, id))
	}

	sort.Slice(result, sortNodesFunc(result))

	k.runOnce = true

	return result, !cmp.Equal(result, k.previous, cmp.AllowUnexported(widgets.TreeNode{}))
}

func (d *dashboardClusterVersionsPane) Render() error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
		clusters []*models.V1ClusterResponse

		gardenerVersions           = pieVersion{}
		firewallControllerVersions = pieVersion{}
	)

	ctx, cancel := context.WithTimeout(context.Background(), dashboardRequestsContextTimeout)
	defer cancel()

	clusters, err := d.cache.Clusters(ctx)
	if err != nil {
		return err
	}

	for _, c := range clusters {
		if c.Status == nil || c.Status.Gardener == nil {
			gardenerVersions["unknown"]++
		} else {
			gardenerVersions[*c.Status.Gardener.Version]++
		}

		if c.FirewallControllerVersion == nil {
			firewallControllerVersions["unknown"]++
		} else {
			if *c.FirewallControllerVersion == "" {
				// auto is default, otherwise some clusters will not be aggregated properly
				c.FirewallControllerVersion = pointer.Pointer("auto")
			}
			firewallControllerVersions[*c.FirewallControllerVersion]++
		}
	}

	processedClusters := len(clusters)
	if processedClusters <= 0 {
		return nil
	}

	data, formatter := gardenerVersions.toValues()
	d.reconciledGardenerVersions.Data = data
	d.reconciledGardenerVersions.LabelFormatter = formatter

	data, formatter = firewallControllerVersions.toValues()
	d.firewallControllerVersions.Data = data
	d.firewallControllerVersions.LabelFormatter = formatter

	toRender := []ui.Drawable{d.reconciledGardenerVersions, d.firewallControllerVersions}

	nodes, changed := d.kubernetesVersionsState.update(clusters)
	if changed {
		d.kubernetesVersions.SetNodes(nodes)
		toRender = append(toRender, d.kubernetesVersions)
	}

	ui.Render(toRender...)

	return nil
}

type dashboardVolumePane struct {
	volumeUsedSpace *widgets.Paragraph

	volumeProtectionState *widgets.BarChart
	volumeState           *widgets.BarChart
	clusterState          *widgets.BarChart
	serverState           *widgets.BarChart

	physicalFree     *widgets.Gauge
	compressionRatio *widgets.Gauge

	sem *semaphore.Weighted

	cache *apiCache
}

func NewDashboardVolumePane(cache *apiCache) *dashboardVolumePane {
	d := &dashboardVolumePane{}

	d.sem = semaphore.NewWeighted(1)

	d.volumeProtectionState = widgets.NewBarChart()
	d.volumeProtectionState.Labels = []string{"Protected", "Degraded", "Read-Only", "N/A", "Unknown"}
	d.volumeProtectionState.Title = "Volume Protection State"
	d.volumeProtectionState.PaddingLeft = 5
	d.volumeProtectionState.BarWidth = 5
	d.volumeProtectionState.BarGap = 10
	d.volumeProtectionState.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed, ui.ColorRed, ui.ColorRed}
	if viper.GetString("color-theme") == "default" {
		d.volumeProtectionState.NumStyles = []ui.Style{{Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}, {Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}}
	}

	d.volumeState = widgets.NewBarChart()
	d.volumeState.Labels = []string{"Available", "Failed", "Unknown", "Other"}
	d.volumeState.Title = "Volume State"
	d.volumeState.PaddingLeft = 5
	d.volumeState.BarWidth = 5
	d.volumeState.BarGap = 10
	d.volumeState.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed, ui.ColorRed, ui.ColorYellow}
	if viper.GetString("color-theme") == "default" {
		d.volumeState.NumStyles = []ui.Style{{Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}}
	}

	d.volumeUsedSpace = widgets.NewParagraph()
	d.volumeUsedSpace.Title = "Volume Infos"
	d.volumeUsedSpace.WrapText = false

	d.physicalFree = widgets.NewGauge()
	d.physicalFree.Title = "Free Physical Space"

	d.compressionRatio = widgets.NewGauge()
	d.compressionRatio.Title = "Compression Ratio"
	d.compressionRatio.BarColor = ui.ColorMagenta

	d.clusterState = widgets.NewBarChart()
	d.clusterState.Labels = []string{"OK", "Warning", "Error", "Other"}
	d.clusterState.Title = "Cluster State"
	d.clusterState.PaddingLeft = 5
	d.clusterState.BarWidth = 5
	d.clusterState.BarGap = 10
	d.clusterState.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed, ui.ColorRed}
	if viper.GetString("color-theme") == "default" {
		d.clusterState.NumStyles = []ui.Style{{Fg: ui.ColorWhite}, {Fg: ui.ColorWhite}, {Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}}
	}

	d.serverState = widgets.NewBarChart()
	d.serverState.Labels = []string{"Enabled", "Disabled", "Failed", "Other"}
	d.serverState.Title = "Server State"
	d.serverState.PaddingLeft = 5
	d.serverState.BarWidth = 5
	d.serverState.BarGap = 10
	d.serverState.BarColors = []ui.Color{ui.ColorGreen, ui.ColorYellow, ui.ColorRed, ui.ColorYellow}
	if viper.GetString("color-theme") == "default" {
		d.serverState.NumStyles = []ui.Style{{Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}, {Fg: ui.ColorBlack}, {Fg: ui.ColorWhite}}
	}

	d.cache = cache
	return d
}

func (d *dashboardVolumePane) Name() string {
	return "Volumes"
}

func (d *dashboardVolumePane) Description() string {
	return "Volume health, for operators also cluster health"
}

func (d *dashboardVolumePane) Resize(x1, y1, x2, y2 int) {
	columnWidth := int(math.Ceil((float64(x2) - (float64(x1))) / 2))
	rowHeight := int(math.Ceil((float64(y2) - (float64(y1))) / 2))

	d.volumeState.SetRect(x1, y1, x1+columnWidth, rowHeight)
	d.volumeProtectionState.SetRect(columnWidth, y1, x2, rowHeight)

	d.volumeUsedSpace.SetRect(x1, rowHeight, x2, rowHeight+3)

	d.physicalFree.SetRect(x1, rowHeight+3, x1+columnWidth, rowHeight+6)
	d.compressionRatio.SetRect(columnWidth, rowHeight+3, x2, rowHeight+6)

	d.clusterState.SetRect(x1, rowHeight+6, x1+columnWidth, y2)
	d.serverState.SetRect(columnWidth, rowHeight+6, x2, y2)
}

func (d *dashboardVolumePane) Render() error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
		clusters []*models.V1StorageClusterInfo
		volumes  []*models.V1VolumeResponse

		volumesProtectionFullyProtected int
		volumesProtectionDegraded       int
		volumesProtectionReadOnly       int
		volumesProtectionNotAvailable   int
		volumesProtectionUnknown        int

		volumesAvailable int
		volumesFailed    int
		volumesUnknown   int
		volumesOther     int

		volumesUsedLogical  int64
		volumesUsedPhysical int64

		clusterStateOK      int
		clusterStateError   int
		clusterStateWarning int
		clusterStateOther   int

		serversEnabled  int
		serversDisabled int
		serversFailed   int
		serversOther    int

		physicalFree     int64
		physicalUsed     int64
		compressionRatio float64
	)

	ctx, cancel := context.WithTimeout(context.Background(), dashboardRequestsContextTimeout)
	defer cancel()

	volumes, err := d.cache.Volumes(ctx)
	if err != nil {
		return err
	}

	clusters, err = d.cache.VolumeClusterInfo(ctx)
	if err != nil {
		var typedErr *volume.ClusterInfoDefault
		if errors.As(err, &typedErr) {
			if typedErr.Code() != http.StatusForbidden {
				return err
			}
			// allow forbidden response, because cluster info is only for provider admins
		} else {
			return err
		}
	}

	for _, v := range volumes {
		if v.State == nil || v.ProtectionState == nil {
			volumesOther++
			volumesProtectionUnknown++
			continue
		}

		switch *v.State {
		case durosv2.Volume_Available.String():
			volumesAvailable++
		case durosv2.Volume_Failed.String():
			volumesFailed++
		case durosv2.Volume_Unknown.String():
			volumesUnknown++
		default:
			volumesOther++
		}

		switch *v.ProtectionState {
		case durosv2.ProtectionStateEnum_FullyProtected.String():
			volumesProtectionFullyProtected++
		case durosv2.ProtectionStateEnum_Degraded.String():
			volumesProtectionDegraded++
		case durosv2.ProtectionStateEnum_ReadOnly.String():
			volumesProtectionReadOnly++
		case durosv2.ProtectionStateEnum_NotAvailable.String():
			volumesProtectionNotAvailable++
		case durosv2.ProtectionStateEnum_Unknown.String():
			volumesProtectionUnknown++
		default:
			volumesProtectionUnknown++
		}

		if v.Statistics != nil {
			if v.Statistics.LogicalUsedStorage != nil {
				volumesUsedLogical += *v.Statistics.LogicalUsedStorage
			}
			if v.Statistics.PhysicalUsedStorage != nil {
				volumesUsedPhysical += *v.Statistics.PhysicalUsedStorage
			}
		}
	}

	d.volumeUsedSpace.Text = fmt.Sprintf("Summed up physical size of volumes: %s, logical size: %s.", helper.HumanizeSize(volumesUsedPhysical), helper.HumanizeSize(volumesUsedLogical))
	ui.Render(d.volumeUsedSpace)

	// for some reason the UI hangs when all values are zero...
	if volumesAvailable > 0 || volumesFailed > 0 || volumesUnknown > 0 || volumesOther > 0 {
		d.volumeState.Data = []float64{float64(volumesAvailable), float64(volumesFailed), float64(volumesUnknown), float64(volumesOther)}
		ui.Render(d.volumeState)
	}

	// for some reason the UI hangs when all values are zero...
	if volumesProtectionFullyProtected > 0 || volumesProtectionDegraded > 0 || volumesProtectionReadOnly > 0 || volumesProtectionNotAvailable > 0 || volumesProtectionUnknown > 0 {
		d.volumeProtectionState.Data = []float64{float64(volumesProtectionFullyProtected), float64(volumesProtectionDegraded), float64(volumesProtectionReadOnly), float64(volumesProtectionNotAvailable), float64(volumesProtectionUnknown)}
		ui.Render(d.volumeProtectionState)
	}

	if len(clusters) == 0 {
		return nil
	}

	for _, c := range clusters {
		if c.Health == nil || c.Health.State == nil {
			clusterStateOther++
			continue
		}

		switch *c.Health.State {
		case durosv2.ClusterHealth_OK.String():
			clusterStateOK++
		case durosv2.ClusterHealth_Error.String():
			clusterStateError++
		case durosv2.ClusterHealth_Warning.String():
			clusterStateWarning++
		case durosv2.ClusterHealth_None.String():
			clusterStateOther++
		default:
			clusterStateOther++
		}

		for _, s := range c.Servers {
			if s.State == nil {
				serversOther++
				continue
			}

			switch *s.State {
			case durosv2.Server_Enabled.String():
				serversEnabled++
			case durosv2.Server_Disabled.String():
				serversDisabled++
			case durosv2.Server_Failed.String():
				serversFailed++
			default:
				serversOther++
			}
		}

		if c.Statistics != nil {
			if c.Statistics.FreePhysicalStorage == nil || c.Statistics.PhysicalUsedStorage == nil || c.Statistics.CompressionRatio == nil {
				continue
			}

			physicalFree += *c.Statistics.FreePhysicalStorage
			physicalUsed += *c.Statistics.PhysicalUsedStorage
			compressionRatio += *c.Statistics.CompressionRatio
		}
	}

	d.compressionRatio.Percent = int((compressionRatio / float64(len(clusters))) * 100)

	totalStorage := float64(physicalFree + physicalUsed)
	d.physicalFree.Percent = int((float64(physicalFree) / totalStorage) * 100)
	if d.physicalFree.Percent < 10 {
		d.physicalFree.BarColor = ui.ColorRed
	} else if d.physicalFree.Percent < 30 {
		d.physicalFree.BarColor = ui.ColorYellow
	} else {
		d.physicalFree.BarColor = ui.ColorGreen
	}

	// for some reason the UI hangs when all values are zero...
	if d.compressionRatio.Percent > 0 && d.physicalFree.Percent > 0 {
		ui.Render(d.compressionRatio, d.physicalFree)
	}

	// for some reason the UI hangs when all values are zero...
	if clusterStateOK > 0 || clusterStateError > 0 || clusterStateWarning > 0 || clusterStateOther > 0 {
		d.clusterState.Data = []float64{float64(clusterStateOK), float64(clusterStateWarning), float64(clusterStateError), float64(clusterStateOther)}
		ui.Render(d.clusterState)
	}

	// for some reason the UI hangs when all values are zero...
	if serversEnabled > 0 || serversDisabled > 0 || serversFailed > 0 || serversOther > 0 {
		d.serverState.Data = []float64{float64(serversEnabled), float64(serversDisabled), float64(serversFailed), float64(serversOther)}
		ui.Render(d.serverState)
	}

	return nil
}

type apiCache struct {
	clusters           *cache.Cache[string, []*models.V1ClusterResponse]
	volumes            *cache.Cache[string, []*models.V1VolumeResponse]
	volumesClusterInfo *cache.Cache[string, []*models.V1StorageClusterInfo]
}

func newCache(cloud *client.CloudAPI, expiration time.Duration, partition, tenant, purpose string) *apiCache {
	return &apiCache{
		clusters: cache.New(expiration, func(ctx context.Context, _ string) ([]*models.V1ClusterResponse, error) {
			resp, err := cloud.Cluster.FindClusters(cluster.NewFindClustersParams().WithBody(&models.V1ClusterFindRequest{
				PartitionID: output.StrDeref(partition),
				Tenant:      output.StrDeref(tenant),
				Purpose:     output.StrDeref(purpose),
			}).WithReturnMachines(pointer.Pointer(false)).WithContext(ctx), nil)
			if err != nil {
				return nil, err
			}
			return resp.Payload, nil
		}),
		volumes: cache.New(expiration, func(ctx context.Context, _ string) ([]*models.V1VolumeResponse, error) {
			resp, err := cloud.Volume.FindVolumes(volume.NewFindVolumesParams().WithBody(&models.V1VolumeFindRequest{
				PartitionID: output.StrDeref(partition),
				TenantID:    output.StrDeref(tenant),
			}).WithContext(ctx), nil)
			if err != nil {
				return nil, err
			}
			return resp.Payload, nil
		}),
		volumesClusterInfo: cache.New(expiration, func(ctx context.Context, _ string) ([]*models.V1StorageClusterInfo, error) {
			resp, err := cloud.Volume.ClusterInfo(volume.NewClusterInfoParams().WithPartitionid(&partition).WithContext(ctx), nil)
			if err != nil {
				return nil, err
			}
			return resp.Payload, nil
		}),
	}
}

func (c *apiCache) Clusters(ctx context.Context) ([]*models.V1ClusterResponse, error) {
	return c.clusters.Get(ctx, "")
}

func (c *apiCache) Volumes(ctx context.Context) ([]*models.V1VolumeResponse, error) {
	return c.volumes.Get(ctx, "")
}

func (c *apiCache) VolumeClusterInfo(ctx context.Context) ([]*models.V1StorageClusterInfo, error) {
	return c.volumesClusterInfo.Get(ctx, "")
}
