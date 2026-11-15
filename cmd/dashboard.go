package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/health"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-lib/pkg/cache"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/healthstatus"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"

	durosv2 "github.com/metal-stack/duros-go/api/duros/v2"
)

const (
	dashboardRequestsContextTimeout = 5 * time.Second
	dashboardHeaderHeight           = 6
)

func newDashboardCmd(c *config) *cobra.Command {
	dashboardCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "shows a live dashboard optimized for operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard(c)
		},
	}

	tabs := dashboardTabs(nil)

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

func runDashboard(c *config) error {
	d, err := NewDashboard(c)
	if err != nil {
		return err
	}

	p := tea.NewProgram(d)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to initialize bubbletea ui: %w", err)
	}

	return nil
}

func dashboardTabs(cache *apiCache) dashboardTabPanes {
	return dashboardTabPanes{
		NewDashboardClusterHealthPane(cache),
		NewDashboardClusterVersionsPane(cache),
		NewDashboardVolumePane(cache),
		NewDashboardClusterListPane(cache),
	}
}

// dashboardTabPane is a single tab of the dashboard. Data is fetched during
// Refresh and rendered to a string during View, so it composes with bubbletea.
// Update may return a command, e.g. to trigger an action such as watching a
// selected cluster.
type dashboardTabPane interface {
	Name() string
	Description() string
	Refresh(ctx context.Context) error
	Update(msg tea.KeyPressMsg) tea.Cmd
	View(width, height int) string
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

// dashboardTickMsg triggers a periodic refresh.
type dashboardTickMsg struct{}

// dashboardRefreshMsg is emitted after an ad-hoc refresh (e.g. when switching tabs).
type dashboardRefreshMsg struct{}

// dashboardSelectClusterMsg requests watching the cluster with the given id.
type dashboardSelectClusterMsg struct {
	id string
}

// dashboardWatchFinishedMsg is emitted after the cluster watch TUI exited.
type dashboardWatchFinishedMsg struct {
	err error
}

// dashboardData holds the shared header state, guarded by a mutex because it is
// written from refresh commands running in a goroutine while View reads it.
type dashboardData struct {
	sync.Mutex

	apiVersion       string
	apiHealth        string
	apiHealthMessage string
	releaseVersion   string
	lastUpdate       time.Time
	lastErr          error
}

// dashboard is our tea.Model for the operational dashboard.
type dashboard struct {
	width, height int
	activeTab     int
	tabs          dashboardTabPanes

	filterTenant    string
	filterPartition string
	filterPurpose   string

	refreshInterval time.Duration

	data  *dashboardData
	cloud *client.CloudAPI

	// watchCluster suspends the dashboard and runs the cluster watch TUI for
	// the given cluster id. It may be nil in tests.
	watchCluster func(id string) error
}

// Init is being called once on boot.
func (d dashboard) Init() tea.Cmd {
	return tea.Batch(d.refreshCmd(), d.tickCmd())
}

// Update handles incoming messages and key events.
func (d dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width, d.height = msg.Width, msg.Height
	case dashboardTickMsg:
		return d, tea.Batch(d.tickCmd(), d.refreshCmd())
	case dashboardRefreshMsg:
		return d, nil
	case dashboardSelectClusterMsg:
		return d, d.watchClusterCmd(msg.id)
	case dashboardWatchFinishedMsg:
		if msg.err != nil {
			d.data.Lock()
			d.data.lastErr = msg.err
			d.data.Unlock()
		}
		return d, d.refreshCmd()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit
		}

		if n, err := strconv.Atoi(msg.String()); err == nil && n >= 1 && n <= len(d.tabs) {
			if d.activeTab != n-1 {
				d.activeTab = n - 1
				return d, d.refreshCmd()
			}
			return d, nil
		}

		// pass the event on to the active tab, which may use it
		return d, d.tabs[d.activeTab].Update(msg)
	}

	return d, nil
}

// dashboardClusterWatchExec adapts an in-process function to the tea.ExecCommand
// interface, so the dashboard can be suspended while the cluster watch TUI runs.
type dashboardClusterWatchExec struct {
	run func() error
}

func (e dashboardClusterWatchExec) Run() error          { return e.run() }
func (e dashboardClusterWatchExec) SetStdin(io.Reader)  {}
func (e dashboardClusterWatchExec) SetStdout(io.Writer) {}
func (e dashboardClusterWatchExec) SetStderr(io.Writer) {}

// watchClusterCmd suspends the dashboard, runs the cluster watch TUI and
// resumes the dashboard once it exits.
func (d dashboard) watchClusterCmd(id string) tea.Cmd {
	if d.watchCluster == nil {
		return nil
	}

	return tea.Exec(dashboardClusterWatchExec{
		run: func() error {
			return d.watchCluster(id)
		},
	}, func(err error) tea.Msg {
		return dashboardWatchFinishedMsg{err: err}
	})
}

// View renders the whole dashboard.
func (d dashboard) View() tea.View {
	width, height := d.width, d.height
	if width <= 0 || height <= 0 {
		width, height = 80, 24
	}

	paneHeight := max(height-dashboardHeaderHeight-2, 4)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		d.headerView(width),
		d.tabsView(),
		d.tabs[d.activeTab].View(width, paneHeight),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#242424")).
			Width(width).
			Render("Switch between tabs with number keys. Press q to quit."),
	)

	v := tea.NewView(body)
	v.AltScreen = true // fullscreen

	return v
}

// headerView renders the status and filter headers side by side.
func (d dashboard) headerView(width int) string {
	d.data.Lock()
	defer d.data.Unlock()

	coloredHealth := d.data.apiHealth
	switch healthstatus.HealthStatus(d.data.apiHealth) {
	case healthstatus.HealthStatusHealthy:
		coloredHealth = lipgloss.NewStyle().Foreground(dashCloseColor).Render(d.data.apiHealth)
	case healthstatus.HealthStatusDegraded, healthstatus.HealthStatusPartiallyUnhealthy:
		coloredHealth = lipgloss.NewStyle().Foreground(dashWarnColor).Render(d.data.apiHealth)
	case healthstatus.HealthStatusUnhealthy:
		label := d.data.apiHealth
		if d.data.apiHealthMessage != "" {
			label = fmt.Sprintf("%s (%s)", d.data.apiHealth, d.data.apiHealthMessage)
		}
		coloredHealth = lipgloss.NewStyle().Foreground(dashBadColor).Render(label)
	}

	versionLine := fmt.Sprintf("cloud-api %s (API Health: %s), cloudctl %s (%s), release %s", d.data.apiVersion, coloredHealth, v.Version, v.GitSHA1, d.data.releaseVersion)

	fetchInfoLine := fmt.Sprintf("Last Update: %s", d.data.lastUpdate.Format("15:04:05"))
	if d.data.lastErr != nil {
		fetchInfoLine += lipgloss.NewStyle().Foreground(dashBadColor).Render(fmt.Sprintf(", Update Error: %s", d.data.lastErr))
	}

	left := lipgloss.JoinVertical(
		lipgloss.Left,
		versionLine,
		fetchInfoLine,
	)

	right := fmt.Sprintf("Tenant=%s\nPartition=%s\nPurpose=%s", d.filterTenant, d.filterPartition, d.filterPurpose)

	leftWidth := max(width-25, 20)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		dashPanel("Cloud Dashboard", left, leftWidth),
		dashPanel("Filters", right, width-leftWidth),
	)
}

// tabsView renders the tab bar.
func (d dashboard) tabsView() string {
	parts := make([]string, 0, len(d.tabs))
	for i, t := range d.tabs {
		label := fmt.Sprintf("(%d) %s", i+1, t.Name())
		if i == d.activeTab {
			parts = append(parts, dashActiveTab.Render(label))
		} else {
			parts = append(parts, dashInactiveTab.Render(label))
		}
	}
	return strings.Join(parts, "  ")
}

// tickCmd schedules the next periodic refresh.
func (d dashboard) tickCmd() tea.Cmd {
	return tea.Tick(d.refreshInterval, func(time.Time) tea.Msg {
		return dashboardTickMsg{}
	})
}

// refreshCmd refreshes all data in the background.
func (d dashboard) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		d.refresh()
		return dashboardRefreshMsg{}
	}
}

// refresh fetches the header info and the currently active tab's data.
func (d dashboard) refresh() {
	var (
		apiVersion       = "unknown"
		apiHealth        = "unknown"
		apiHealthMessage string
		releaseVersion   = "unknown"

		lastErr error
	)

	ctx, cancel := context.WithTimeout(context.Background(), dashboardRequestsContextTimeout)
	defer cancel()

	infoResp, err := d.cloud.Version.Info(version.NewInfoParams().WithContext(ctx), nil)
	if err != nil {
		lastErr = err
	} else {
		if infoResp.Payload.Version != nil {
			apiVersion = *infoResp.Payload.Version
		}
		releaseVersion = infoResp.Payload.ReleaseVersion
	}

	if lastErr == nil {
		healthResp, err := d.cloud.Health.Health(health.NewHealthParams().WithContext(ctx), nil)
		if err != nil {
			if r, ok := errors.AsType[*health.HealthInternalServerError](err); ok {
				healthResp = health.NewHealthOK()
				healthResp.Payload = r.Payload
			} else {
				lastErr = err
			}
		}

		if healthResp != nil && healthResp.Payload != nil {
			if healthResp.Payload.Status != nil {
				apiHealth = *healthResp.Payload.Status
			}
			if healthResp.Payload.Message != nil {
				apiHealthMessage = *healthResp.Payload.Message
			}
		}
	}

	if lastErr == nil {
		lastErr = d.tabs[d.activeTab].Refresh(ctx)
	}

	d.data.Lock()
	d.data.apiVersion = apiVersion
	d.data.apiHealth = apiHealth
	d.data.apiHealthMessage = apiHealthMessage
	d.data.releaseVersion = releaseVersion
	d.data.lastUpdate = time.Now()
	d.data.lastErr = lastErr
	d.data.Unlock()
}

func NewDashboard(c *config) (*dashboard, error) {
	if err := dashboardApplyTheme(viper.GetString("color-theme")); err != nil {
		return nil, err
	}

	d := &dashboard{
		width:           80,
		height:          24,
		filterTenant:    viper.GetString("tenant"),
		filterPartition: viper.GetString("partition"),
		filterPurpose:   viper.GetString("purpose"),
		refreshInterval: viper.GetDuration("refresh-interval"),
		data:            &dashboardData{},
		cloud:           c.cloud,
		watchCluster: func(id string) error {
			return c.clusterWatch([]string{id})
		},
	}

	cache := newCache(c.cloud,
		viper.GetDuration("refresh-interval"),
		viper.GetString("partition"),
		viper.GetString("tenant"),
		viper.GetString("purpose"),
	)

	d.tabs = dashboardTabs(cache)

	if viper.IsSet("initial-tab") {
		initialPanelIndex, err := d.tabs.FindIndexByName(viper.GetString("initial-tab"))
		if err != nil {
			return nil, err
		}
		d.activeTab = initialPanelIndex
	}

	return d, nil
}

type dashboardClusterError struct {
	ClusterName  string
	ErrorMessage string
	Time         time.Time
}

var (
	clusterHealthLabels = []string{"Succeeded", "Progressing", "Unhealthy"}
	clusterHealthColors = []dashColor{dashCloseColor, dashWarnColor, dashBadColor}
)

type dashboardClusterHealthPane struct {
	mu  sync.Mutex
	sem *semaphore.Weighted

	clusterHealth []float64

	apiPercent     int
	controlPercent int
	nodesPercent   int
	systemPercent  int

	problems   [][]string
	lastErrors [][]string

	cache *apiCache
}

func NewDashboardClusterHealthPane(cache *apiCache) *dashboardClusterHealthPane {
	return &dashboardClusterHealthPane{
		sem:   semaphore.NewWeighted(1),
		cache: cache,
	}
}

func (d *dashboardClusterHealthPane) Name() string {
	return "ClusterHealth"
}

func (d *dashboardClusterHealthPane) Description() string {
	return "Cluster health and issues"
}

func (d *dashboardClusterHealthPane) Update(tea.KeyPressMsg) tea.Cmd { return nil }

func (d *dashboardClusterHealthPane) Refresh(ctx context.Context) error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
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

	sort.Slice(clusterErrors, func(i, j int) bool {
		return clusterErrors[i].Time.After(clusterErrors[j].Time)
	})
	problemRows := make([][]string, 0, len(clusterErrors))
	for _, e := range clusterErrors {
		problemRows = append(problemRows, []string{e.ClusterName, e.ErrorMessage})
	}

	sort.Slice(lastErrors, func(i, j int) bool {
		return lastErrors[i].Time.After(lastErrors[j].Time)
	})
	lastErrorRows := make([][]string, 0, len(lastErrors))
	for _, e := range lastErrors {
		lastErrorRows = append(lastErrorRows, []string{e.ClusterName, e.ErrorMessage})
	}

	d.mu.Lock()
	d.clusterHealth = []float64{float64(succeeded), float64(processing), float64(unhealthy)}
	d.apiPercent = apiOK * 100 / processedClusters
	d.controlPercent = controlOK * 100 / processedClusters
	d.nodesPercent = nodesOK * 100 / processedClusters
	d.systemPercent = systemOK * 100 / processedClusters
	d.problems = problemRows
	d.lastErrors = lastErrorRows
	d.mu.Unlock()

	return nil
}

func (d *dashboardClusterHealthPane) View(width, height int) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	leftWidth := min(50, width/2)
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := width - leftWidth

	left := dashPanel("Cluster Operation",
		dashBarChart(clusterHealthLabels, d.clusterHealth, clusterHealthColors, leftWidth-6), leftWidth)

	gauges := strings.Join([]string{
		dashGauge("API", d.apiPercent, dashPercentColor(d.apiPercent), rightWidth-6),
		dashGauge("Control", d.controlPercent, dashPercentColor(d.controlPercent), rightWidth-6),
		dashGauge("Nodes", d.nodesPercent, dashPercentColor(d.nodesPercent), rightWidth-6),
		dashGauge("System", d.systemPercent, dashPercentColor(d.systemPercent), rightWidth-6),
	}, "\n")
	right := dashPanel("Cluster Status", gauges, rightWidth)

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	tableHeight := max((height-lipgloss.Height(top)-4)/2, 3)

	problems := addCaption(dashTable([]string{"CLUSTER", "MESSAGE"}, d.problems, width-4, tableHeight), "Cluster Problems")
	lastErrors := addCaption(dashTable([]string{"CLUSTER", "MESSAGE"}, d.lastErrors, width-4, tableHeight), "Last Errors")

	return lipgloss.JoinVertical(lipgloss.Left, top, problems, lastErrors)
}

type dashboardClusterVersionsPane struct {
	mu  sync.Mutex
	sem *semaphore.Weighted

	gardenerVersions []pieEntry
	firewallVersions []pieEntry

	kubernetesVersionsState *kubernetesVersions
	selected                int
	scroll                  int

	cache *apiCache
}

func NewDashboardClusterVersionsPane(cache *apiCache) *dashboardClusterVersionsPane {
	return &dashboardClusterVersionsPane{
		sem:                     semaphore.NewWeighted(1),
		kubernetesVersionsState: &kubernetesVersions{},
		cache:                   cache,
	}
}

func (d *dashboardClusterVersionsPane) Name() string {
	return "ClusterVersions"
}

func (d *dashboardClusterVersionsPane) Description() string {
	return "Cluster version information"
}

func (d *dashboardClusterVersionsPane) Update(msg tea.KeyPressMsg) tea.Cmd {
	d.mu.Lock()
	defer d.mu.Unlock()

	rows := flattenVersionTree(d.kubernetesVersionsState.previous)
	if len(rows) == 0 {
		return nil
	}

	switch msg.String() {
	case "down", "j":
		if d.selected < len(rows)-1 {
			d.selected++
		}
	case "up", "k":
		if d.selected > 0 {
			d.selected--
		}
	case "pgdown", "f":
		d.selected += 10
	case "pgup", "b":
		d.selected -= 10
	case "home", "g":
		d.selected = 0
	case "end", "G":
		d.selected = len(rows) - 1
	case "enter", " ", "space":
		rows[d.selected].node.expanded = !rows[d.selected].node.expanded
	case "right", "l":
		rows[d.selected].node.expanded = true
	case "left", "h":
		rows[d.selected].node.expanded = false
	case "E":
		setVersionTreeExpanded(d.kubernetesVersionsState.previous, true)
	case "C":
		setVersionTreeExpanded(d.kubernetesVersionsState.previous, false)
	}

	rows = flattenVersionTree(d.kubernetesVersionsState.previous)
	if len(rows) == 0 {
		d.selected = 0
		return nil
	}
	if d.selected < 0 {
		d.selected = 0
	}
	if d.selected > len(rows)-1 {
		d.selected = len(rows) - 1
	}

	return nil
}

func (d *dashboardClusterVersionsPane) Refresh(ctx context.Context) error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
		gardenerVersions           = pieVersion{}
		firewallControllerVersions = pieVersion{}
	)

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
				c.FirewallControllerVersion = new("auto")
			}
			firewallControllerVersions[*c.FirewallControllerVersion]++
		}
	}

	d.mu.Lock()
	nodes, _ := d.kubernetesVersionsState.update(clusters)
	d.gardenerVersions = gardenerVersions.entries()
	d.firewallVersions = firewallControllerVersions.entries()
	rows := flattenVersionTree(nodes)
	if len(rows) == 0 {
		d.selected = 0
	} else if d.selected > len(rows)-1 {
		d.selected = len(rows) - 1
	}
	d.mu.Unlock()

	return nil
}

func (d *dashboardClusterVersionsPane) View(width, height int) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	leftWidth := width - 40
	if leftWidth < 20 {
		leftWidth = width / 2
	}
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := width - leftWidth

	left := dashPanel("Kubernetes Versions", d.renderTree(height-2), leftWidth)
	half := max((height-4)/2, 2)
	right := lipgloss.JoinVertical(lipgloss.Left,
		dashPanel("Reconciled Gardener Versions", dashMaxLines(dashPie(d.gardenerVersions, rightWidth-6), half), rightWidth),
		dashPanel("Firewall-Controller Versions", dashMaxLines(dashPie(d.firewallVersions, rightWidth-6), half), rightWidth),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (d *dashboardClusterVersionsPane) renderTree(height int) string {
	rows := flattenVersionTree(d.kubernetesVersionsState.previous)
	if len(rows) == 0 {
		return ""
	}

	if height < 1 {
		height = 1
	}

	if d.selected < d.scroll {
		d.scroll = d.selected
	}
	if d.selected >= d.scroll+height {
		d.scroll = d.selected - height + 1
	}
	if d.scroll < 0 {
		d.scroll = 0
	}
	if maxScroll := len(rows) - height; maxScroll < 0 {
		d.scroll = 0
	} else if d.scroll > maxScroll {
		d.scroll = maxScroll
	}

	var b strings.Builder
	for i := d.scroll; i < len(rows) && i < d.scroll+height; i++ {
		r := rows[i]

		marker := "  "
		if len(r.node.kids) > 0 {
			if r.node.expanded {
				marker = "▾ "
			} else {
				marker = "▸ "
			}
		}

		line := strings.Repeat("  ", r.depth) + marker + r.node.value
		if i == d.selected {
			line = dashSelected.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

type pieVersion map[string]int

func (p pieVersion) total() int {
	var result int
	for _, amount := range p {
		result += amount
	}
	return result
}

func (p pieVersion) entries() []pieEntry {
	type version struct {
		parsedVersion *semver.Version
		rawVersion    string
		count         int
		percentage    float64
	}

	total := p.total()
	if total == 0 {
		return nil
	}

	var versions []version

	for rawVersion, count := range p {
		v, _ := semver.NewVersion(rawVersion)

		versions = append(versions, version{
			parsedVersion: v,
			rawVersion:    rawVersion,
			count:         count,
			percentage:    float64(count) / float64(total),
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

	result := make([]pieEntry, 0, len(versions))
	for _, v := range versions {
		result = append(result, pieEntry{
			Label:   v.rawVersion,
			Count:   v.count,
			Percent: v.percentage,
		})
	}

	return result
}

// versionNode is a node in the kubernetes version tree. The expanded state is
// carried over between refreshes by reusing matching nodes.
type versionNode struct {
	value    string
	expanded bool
	kids     []*versionNode
}

type versionTreeRow struct {
	node  *versionNode
	depth int
}

func flattenVersionTree(nodes []*versionNode) []versionTreeRow {
	var rows []versionTreeRow

	var walk func(ns []*versionNode, depth int)
	walk = func(ns []*versionNode, depth int) {
		for _, n := range ns {
			rows = append(rows, versionTreeRow{node: n, depth: depth})
			if n.expanded {
				walk(n.kids, depth+1)
			}
		}
	}
	walk(nodes, 0)

	return rows
}

func setVersionTreeExpanded(nodes []*versionNode, expanded bool) {
	for _, n := range nodes {
		if len(n.kids) > 0 {
			n.expanded = expanded
		}
		setVersionTreeExpanded(n.kids, expanded)
	}
}

type kubernetesVersions struct {
	previous []*versionNode
}

func (k *kubernetesVersions) update(clusters []*models.V1ClusterResponse) ([]*versionNode, bool) {
	var (
		result []*versionNode

		sortNodesFunc = func(nodes []*versionNode) func(i, j int) bool {
			return func(i, j int) bool {
				iParsed, _ := semver.NewVersion(nodes[i].value)
				jParsed, _ := semver.NewVersion(nodes[j].value)

				if iParsed == nil && jParsed == nil {
					return nodes[i].value < nodes[j].value
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

		findNode func(nodes []*versionNode, name string) *versionNode

		ensureNode func(name string) *versionNode

		ensureChildNode func(parent *versionNode, childName string) *versionNode
	)

	findNode = func(nodes []*versionNode, name string) *versionNode {
		for _, n := range nodes {
			if n.value == name {
				return n
			}
			if found := findNode(n.kids, name); found != nil {
				return found
			}
		}
		return nil
	}

	ensureNode = func(name string) *versionNode {
		if existing := findNode(result, name); existing != nil {
			return existing
		}

		if previous := findNode(k.previous, name); previous != nil {
			result = append(result, previous)
			return previous
		}

		newNode := &versionNode{
			value: name,
		}

		result = append(result, newNode)

		sort.Slice(result, sortNodesFunc(result))

		return newNode
	}

	ensureChildNode = func(parent *versionNode, childName string) *versionNode {
		if existing := findNode(parent.kids, childName); existing != nil {
			return existing
		}

		newNode := &versionNode{
			value: childName,
		}

		parent.kids = append(parent.kids, newNode)

		sort.Slice(parent.kids, sortNodesFunc(parent.kids))

		return newNode
	}

	majorMinorAmounts := map[string]int{}
	versionAmounts := map[string]int{}

	for _, c := range clusters {

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

	changed := !cmp.Equal(result, k.previous, cmp.AllowUnexported(versionNode{}))
	k.previous = result

	return result, changed
}

var (
	volumeStateLabels      = []string{"Available", "Failed", "Unknown", "Other"}
	volumeStateColors      = []dashColor{dashCloseColor, dashBadColor, dashBadColor, dashWarnColor}
	volumeProtectionLabels = []string{"Protected", "Degraded", "Read-Only", "N/A", "Unknown"}
	volumeProtectionColors = []dashColor{dashCloseColor, dashWarnColor, dashBadColor, dashBadColor, dashBadColor}
	clusterStateLabels     = []string{"OK", "Warning", "Error", "Other"}
	clusterStateColors     = []dashColor{dashCloseColor, dashWarnColor, dashBadColor, dashBadColor}
	serverStateLabels      = []string{"Enabled", "Disabled", "Failed", "Other"}
	serverStateColors      = []dashColor{dashCloseColor, dashWarnColor, dashBadColor, dashWarnColor}
)

type dashboardVolumePane struct {
	mu  sync.Mutex
	sem *semaphore.Weighted

	usedSpaceText string

	volumeState     []float64
	protectionState []float64
	clusterState    []float64
	serverState     []float64

	physicalFreePercent     int
	physicalFreeColor       dashColor
	compressionRatioPercent int

	cache *apiCache
}

func NewDashboardVolumePane(cache *apiCache) *dashboardVolumePane {
	return &dashboardVolumePane{
		sem:               semaphore.NewWeighted(1),
		cache:             cache,
		physicalFreeColor: dashCloseColor,
	}
}

func (d *dashboardVolumePane) Name() string {
	return "Volumes"
}

func (d *dashboardVolumePane) Description() string {
	return "Volume health, for operators also cluster health"
}

func (d *dashboardVolumePane) Update(tea.KeyPressMsg) tea.Cmd { return nil }

func (d *dashboardVolumePane) Refresh(ctx context.Context) error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
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

	volumes, err := d.cache.Volumes(ctx)
	if err != nil {
		return err
	}

	clusterInfos, err := d.cache.VolumeClusterInfo(ctx)
	if err != nil {
		if typedErr, ok := errors.AsType[*volume.ClusterInfoDefault](err); ok {
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

	usedSpaceText := fmt.Sprintf("Summed up physical size of volumes: %s, logical size: %s.", helper.HumanizeSize(volumesUsedPhysical), helper.HumanizeSize(volumesUsedLogical))

	for _, c := range clusterInfos {
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

	physicalFreePercent := 0
	if totalStorage := float64(physicalFree + physicalUsed); totalStorage > 0 {
		physicalFreePercent = int((float64(physicalFree) / totalStorage) * 100)
	}

	physicalFreeColor := dashCloseColor
	switch {
	case physicalFreePercent < 10:
		physicalFreeColor = dashBadColor
	case physicalFreePercent < 30:
		physicalFreeColor = dashWarnColor
	}

	compressionRatioPercent := 0
	if len(clusterInfos) > 0 {
		compressionRatioPercent = int((compressionRatio / float64(len(clusterInfos))) * 100)
	}

	d.mu.Lock()
	d.usedSpaceText = usedSpaceText
	d.volumeState = []float64{float64(volumesAvailable), float64(volumesFailed), float64(volumesUnknown), float64(volumesOther)}
	d.protectionState = []float64{float64(volumesProtectionFullyProtected), float64(volumesProtectionDegraded), float64(volumesProtectionReadOnly), float64(volumesProtectionNotAvailable), float64(volumesProtectionUnknown)}
	d.clusterState = []float64{float64(clusterStateOK), float64(clusterStateWarning), float64(clusterStateError), float64(clusterStateOther)}
	d.serverState = []float64{float64(serversEnabled), float64(serversDisabled), float64(serversFailed), float64(serversOther)}
	d.physicalFreePercent = physicalFreePercent
	d.physicalFreeColor = physicalFreeColor
	d.compressionRatioPercent = compressionRatioPercent
	d.mu.Unlock()

	return nil
}

func (d *dashboardVolumePane) View(width, height int) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	columnWidth := max(width/2, 20)

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		dashPanel("Volume State", dashBarChart(volumeStateLabels, d.volumeState, volumeStateColors, columnWidth-6), columnWidth),
		dashPanel("Volume Protection State", dashBarChart(volumeProtectionLabels, d.protectionState, volumeProtectionColors, width-columnWidth-6), width-columnWidth),
	)

	info := dashPanel("Volume Infos", d.usedSpaceText, width)

	gauges := lipgloss.JoinHorizontal(lipgloss.Top,
		dashPanel("Free Physical Space", dashGauge("Free", d.physicalFreePercent, d.physicalFreeColor, columnWidth-6), columnWidth),
		dashPanel("Compression Ratio", dashGauge("Ratio", d.compressionRatioPercent, lipgloss.Color("#da70d6"), width-columnWidth-6), width-columnWidth),
	)

	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		dashPanel("Cluster State", dashBarChart(clusterStateLabels, d.clusterState, clusterStateColors, columnWidth-6), columnWidth),
		dashPanel("Server State", dashBarChart(serverStateLabels, d.serverState, serverStateColors, width-columnWidth-6), width-columnWidth),
	)

	return lipgloss.JoinVertical(lipgloss.Left, top, info, gauges, bottom)
}

// clusterListItem is a pre-extracted row of the cluster list pane.
type clusterListItem struct {
	id         string
	name       string
	tenant     string
	kubernetes string
	status     string
}

type dashboardClusterListPane struct {
	mu  sync.Mutex
	sem *semaphore.Weighted

	clusters []clusterListItem
	selected int
	scroll   int

	cache *apiCache
}

func NewDashboardClusterListPane(cache *apiCache) *dashboardClusterListPane {
	return &dashboardClusterListPane{
		sem:   semaphore.NewWeighted(1),
		cache: cache,
	}
}

func (d *dashboardClusterListPane) Name() string {
	return "Clusters"
}

func (d *dashboardClusterListPane) Description() string {
	return "All clusters, press enter to watch the selected cluster"
}

func (d *dashboardClusterListPane) Update(msg tea.KeyPressMsg) tea.Cmd {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch msg.String() {
	case "down", "j":
		if d.selected < len(d.clusters)-1 {
			d.selected++
		}
	case "up", "k":
		if d.selected > 0 {
			d.selected--
		}
	case "pgdown", "f":
		d.selected += 10
	case "pgup", "b":
		d.selected -= 10
	case "home", "g":
		d.selected = 0
	case "end", "G":
		d.selected = len(d.clusters) - 1
	case "enter", " ", "space":
		if d.selected >= 0 && d.selected < len(d.clusters) {
			id := d.clusters[d.selected].id
			if id != "" {
				return func() tea.Msg {
					return dashboardSelectClusterMsg{id: id}
				}
			}
		}
	}

	if d.selected < 0 {
		d.selected = 0
	}
	if d.selected > len(d.clusters)-1 {
		d.selected = len(d.clusters) - 1
	}

	return nil
}

func (d *dashboardClusterListPane) Refresh(ctx context.Context) error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	clusters, err := d.cache.Clusters(ctx)
	if err != nil {
		return err
	}

	items := make([]clusterListItem, 0, len(clusters))
	for _, c := range clusters {
		status := "unknown"
		if c.Status != nil && c.Status.LastOperation != nil && c.Status.LastOperation.State != nil && *c.Status.LastOperation.State != "" {
			status = *c.Status.LastOperation.State
		}

		items = append(items, clusterListItem{
			id:         pointer.SafeDeref(c.ID),
			name:       pointer.SafeDeref(c.Name),
			tenant:     pointer.SafeDeref(c.Tenant),
			kubernetes: pointer.SafeDeref(pointer.SafeDeref(c.Kubernetes).Version),
			status:     status,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].name == items[j].name {
			return items[i].id < items[j].id
		}
		return items[i].name < items[j].name
	})

	d.mu.Lock()
	d.clusters = items
	if len(items) == 0 {
		d.selected = 0
	} else if d.selected > len(items)-1 {
		d.selected = len(items) - 1
	}
	d.mu.Unlock()

	return nil
}

func (d *dashboardClusterListPane) View(width, height int) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	// the rendered table adds borders, a header and a header separator line
	const tableOverhead = 4
	showHint := height-tableOverhead-1 >= 1
	visibleHeight := height - tableOverhead
	if showHint {
		visibleHeight--
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	if d.selected < d.scroll {
		d.scroll = d.selected
	}
	if d.selected >= d.scroll+visibleHeight {
		d.scroll = d.selected - visibleHeight + 1
	}
	if maxScroll := len(d.clusters) - visibleHeight; maxScroll < 0 {
		d.scroll = 0
	} else if d.scroll > maxScroll {
		d.scroll = maxScroll
	}
	if d.scroll < 0 {
		d.scroll = 0
	}

	end := min(d.scroll+visibleHeight, len(d.clusters))
	visible := d.clusters[d.scroll:end]

	rows := make([][]string, 0, len(visible))
	for _, c := range visible {
		rows = append(rows, []string{c.name, c.id, c.tenant, c.kubernetes, c.status})
	}

	table := dashSelectableTable(
		[]string{"NAME", "ID", "TENANT", "KUBERNETES", "STATUS"},
		rows,
		d.selected-d.scroll,
		width-4,
	)

	if !showHint {
		return table
	}

	hint := dashInactiveTab.Render("Use arrow keys to navigate, press enter to watch the selected cluster.")

	return lipgloss.JoinVertical(lipgloss.Left, table, hint)
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
				PartitionID: pointer.PointerOrNil(partition),
				Tenant:      pointer.PointerOrNil(tenant),
				Purpose:     pointer.PointerOrNil(purpose),
			}).WithReturnMachines(new(false)).WithContext(ctx), nil)
			if err != nil {
				return nil, err
			}
			return resp.Payload, nil
		}),
		volumes: cache.New(expiration, func(ctx context.Context, _ string) ([]*models.V1VolumeResponse, error) {
			resp, err := cloud.Volume.FindVolumes(volume.NewFindVolumesParams().WithBody(&models.V1VolumeFindRequest{
				PartitionID: pointer.PointerOrNil(partition),
				TenantID:    pointer.PointerOrNil(tenant),
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
