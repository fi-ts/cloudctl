package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/v"
)

const (
	// FIXME: make tickDelay customizable
	tickDelay = time.Second * 5
)

// fetchShoot  returns an  updated cluster  struct, the  func will  be
// defined during watchData instantiation,  because it needs access to
// config resources
type fetchShoot func() (*cluster.FindClusterOK, error)

// watchData  stores cluster  data, it  is  attached to  the model  as
// pointer to avoid ELM copy-by-value issues
type watchData struct {
	sync.Mutex

	update fetchShoot
	ticks  int
	err    error
	shoot  *cluster.FindClusterOK
}

// nodes returns a slice of cluster worker nodes
func (wd *watchData) nodes() []*models.ModelsV1MachineResponse {
	ms := wd.shoot.Payload.Machines
	ms = append(ms, wd.shoot.Payload.Firewalls...)
	return ms
}

// nodecount returns the number of nodes
func (wd *watchData) nodecount() string {
	ms := wd.shoot.Payload.Machines
	ms = append(ms, wd.shoot.Payload.Firewalls...)
	return strconv.Itoa(len(ms))
}

// conditions returns a list of cluster status conditions
func (wd *watchData) conditions() []*models.V1beta1Condition {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.Conditions
	}

	return nil
}

// seed returns the seed cluster
func (wd *watchData) seed() string {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.SeedName
	}

	return ""
}

// lastupdate returns the timestamp when the cluster has been updated last
func (wd *watchData) lastupdate() string {
	lastReconciliation := ""

	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil && wd.shoot.Payload.Status.LastOperation.LastUpdateTime != nil {

		lastUpdate, err := time.Parse(time.RFC3339, *wd.shoot.Payload.Status.LastOperation.LastUpdateTime)
		if err != nil {
			lastReconciliation = "unknown"
		} else {
			lastReconciliation = helper.HumanizeDuration(time.Since(lastUpdate))
		}
	}

	return lastReconciliation
}

// lastOperation returns the last cluster operation
func (wd *watchData) lastOperation() *models.V1beta1LastOperation {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.LastOperation
	}

	return nil
}

// lastErrors returns a slice of last cluster errors
func (wd *watchData) lastErrors() []*models.V1beta1LastError {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.LastErrors
	}

	return nil
}

func (wd *watchData) size() string {
	autoScaleMin := int32(0)
	autoScaleMax := int32(0)

	for _, w := range wd.shoot.Payload.Workers {
		autoScaleMin += *w.Minimum
		autoScaleMax += *w.Maximum
	}

	currentMachines := "x"

	if wd.shoot.Payload.Machines != nil {
		currentMachines = fmt.Sprintf("%d", len(wd.shoot.Payload.Machines))
	}

	return fmt.Sprintf("%d≤%s≤%d", autoScaleMin, currentMachines, autoScaleMax)
}

func (wd *watchData) id() string {
	return *wd.shoot.Payload.ID
}

func (wd *watchData) name() string {
	return *wd.shoot.Payload.Name
}

func (wd *watchData) tenant() string {
	return *wd.shoot.Payload.Tenant
}

func (wd *watchData) version() string {
	return *wd.shoot.Payload.Kubernetes.Version
}

func (wd *watchData) partition() string {
	return *wd.shoot.Payload.PartitionID
}

func (wd *watchData) operation() string {
	return *wd.shoot.Payload.Status.LastOperation.State
}

func (wd *watchData) progress() int32 {
	return *wd.shoot.Payload.Status.LastOperation.Progress
}

func (wd *watchData) progressType() string {
	return *wd.shoot.Payload.Status.LastOperation.Type
}

// our key bindings, not [yet] customizable
type KeyMap struct {
	Quit        key.Binding
	ToggleNodes key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	ToggleNodes: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "toggle node list"),
	),
}

// watch is our tea.Model for bubbletea ELM
type watch struct {
	data          *watchData
	width, height int
	nodes         bool // show or hide node list
}

// TickMsg is being used as ELM Tick  cmd to do regular work, which in
// our case is to fetch cluster data
type TickMsg time.Time

// tickInit sends the first tick
func tickInit() tea.Cmd {
	return tea.Every(tickDelay, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init is being called once on boot
func (w watch) Init() tea.Cmd {
	// start the tick timer
	return tickInit()
}

// Update is  being called  regularly and if  an event  occurs. Update
// data and catch keys here.
func (w watch) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TickMsg:
		// Asynchronously update shoot data  and return a new tick msg
		// for the  next  loop.  This is  important,  to  keep the  UI
		// responsive, otherwise it would hang during API fetches.
		return w, tea.Every(tickDelay, func(t time.Time) tea.Msg {
			w.updateShoot()

			return TickMsg(t)
		})

	case tea.WindowSizeMsg:
		w.updateDimensions(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return w, tea.Quit
		case key.Matches(msg, DefaultKeyMap.ToggleNodes):
			w.nodes = !w.nodes
		}
	}

	return w, cmd
}

// View is being called regularly to [re-]render the ui
func (w watch) View() tea.View {
	body := strings.Builder{}

	if w.data.err != nil {
		body.WriteString(fmt.Sprint(w.data.err))
	} else if w.data.shoot != nil {
		nodes := w.nodeRowView()
		help := w.helpView()
		height := w.canvasHeight(lipgloss.Height(nodes), lipgloss.Height(help))

		top := canvas.Height(height).Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				w.topRowView(),
				w.errorRowView(),
				w.conditionRowView(),
			))

		if w.nodes {
			body.WriteString(
				lipgloss.JoinVertical(
					lipgloss.Center,
					top,
					w.nodeRowView(),
					w.helpView()))
		} else {
			body.WriteString(
				lipgloss.JoinVertical(
					lipgloss.Center,
					top,
					w.helpView()))
		}
	}

	v := tea.NewView(body.String())
	v.AltScreen = true // fulscreen

	return v
}

// canvasHeight returns the  height left for the main  canvas, so that
// there's enough room for nodes+help
func (w *watch) canvasHeight(nodesHeight, helpHeight int) int {
	if w.nodes {
		return w.height - nodesHeight - helpHeight
	}

	return w.height - helpHeight
}

// helpView returns the rendered help line
func (w *watch) helpView() string {
	keys := strings.Builder{}

	keys.WriteString(helpStyle.Render("Keys: "))

	for _, key := range []key.Binding{DefaultKeyMap.ToggleNodes, DefaultKeyMap.Quit} {
		keys.WriteString(keyStyle.Render(strings.TrimSpace(key.Help().Key)))
		keys.WriteString(helpStyle.Render(": "))
		keys.WriteString(helpStyle.Render(key.Help().Desc))
	}

	return keyCanvas.Width(w.width).
		Render(keys.String())
}

// errorRowView returns the rendered last error row table
func (w *watch) errorRowView() string {
	lasterrs := w.data.lastErrors()
	lastop := w.data.lastOperation()

	view := errorRow.ClearRows().
		Width(w.width)

	view.Row(
		*lastop.LastUpdateTime +
			" [" + *lastop.Type + "]: " +
			*lastop.Description,
	)

	for _, err := range lasterrs {
		view.Row(
			err.LastUpdateTime +
				" [" + err.TaskID + "]: " +
				*err.Description,
		)
	}

	return addCaption(view.Render(), "Last Events")
}

// nodeRowView returns the rendered node row table
func (w *watch) nodeRowView() string {
	view := nodeRow.ClearRows().
		Width(w.width).
		Headers("HOSTNAME", "STARTED", "AGE", "LAST EVENT", "WHEN", "IMAGE", "RACK")

	for _, node := range w.data.nodes() {
		alloc := node.Allocation
		if alloc == nil {
			continue
		}

		hostname := pointer.SafeDeref(alloc.Hostname)

		image := ""
		if alloc.Image != nil {
			image = pointer.SafeDeref(alloc.Image.ID)
		}

		started := pointer.SafeDeref(alloc.Created)
		age := ""
		format := "2006-01-02T15:04:05.999Z"
		created, err := time.Parse(format, *alloc.Created)
		if err != nil {
			fmt.Printf("unable to parse created time:%s", err)
			os.Exit(1)
		}
		if alloc.Created != nil && !created.IsZero() {
			started = created.Format(time.RFC3339)
			age = output.HumanizeDuration(time.Since(created))
		}

		lastEvent := ""
		lastEventTime, err := time.Parse(format, node.Events.LastEventTime)
		if err != nil {
			fmt.Printf("unable to parse lastevent time:%s", err)
			os.Exit(1)
		}
		when := ""
		if len(node.Events.Log) > 0 {
			since := time.Since(lastEventTime)
			when = output.HumanizeDuration(since)
			lastEvent = *node.Events.Log[0].Event
		}

		view.Row(
			hostname,
			started,
			age,
			lastEvent,
			when,
			image,
			node.Rackid,
		)
	}

	return addCaption(view.Render(), "Cluster Nodes")
}

// conditionRowView returns the rendered condition row table
func (w *watch) conditionRowView() string {
	view := conditionRow.ClearRows().
		Width(w.width).
		Headers("TIME", "LOG MESSAGE", "REASON", "TYPE")

	for _, cond := range w.data.conditions() {
		view.Row(
			*cond.LastTransitionTime,
			*cond.Message,
			*cond.Reason,
			*cond.Type,
		)
	}

	return addCaption(view.Render(), "Cluster Logs")
}

// topRowView returns the rendered top row table
func (w *watch) topRowView() string {
	topRow.ClearRows().
		Width(w.width).
		Row(
			strings.Join([]string{
				renderTopColumn("     ID", w.data.id()),
				renderTopColumn("Cluster", w.data.name()),
				renderTopColumn("Cluster", w.data.seed()),
				renderTopColumn(" Tenant", w.data.tenant()),
			}, "\n"),

			strings.Join([]string{
				renderTopColumn(" K8S Version", w.data.version()),
				renderTopColumn("      Client", v.Version),
				renderTopColumn("   Partition", w.data.partition()),
				renderTopColumn("Cluster Size", w.data.size()),
			}, "\n"),

			strings.Join([]string{
				renderTopColumn("    Nodes", w.data.nodecount()),
				renderTopColumn("  Updated", w.data.lastupdate()+" ago"),
				renderTopColumn("Operation", w.data.operation()),
				renderTopColumn(" Progress", renderProgress(w.data.progress())),
			}, "\n"),
		)

	return addCaption(topRow.Render(), "Cluster Status")
}

// updateDimensions stores current terminal size
func (w *watch) updateDimensions(width, height int) {
	w.width = width
	w.height = height
}

// updateShoot fetches current cluster data  from the API and updates
// the pointers, it runs on every tick (see TickMsg above)
func (w watch) updateShoot() watch {
	w.data.Lock()
	defer w.data.Unlock()

	shoot, err := w.data.update()

	if err != nil {
		w.data.err = err
		return w
	}

	w.data.shoot = shoot
	w.data.ticks++
	w.data.err = nil

	return w
}

// clusterWatch initializes the TUI and boots it
func (c *config) clusterWatch(args []string) error {
	ci, err := c.clusterID("machines", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)

	w := watch{
		data: &watchData{
			update: func() (*cluster.FindClusterOK, error) {
				shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
				if err != nil {
					return nil, err
				}

				return shoot, nil
			}},
	}

	// run once initially
	w.updateShoot()

	p := tea.NewProgram(w)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to initialize bubbletea ui: %w", err)
	}

	return nil
}

// renderTopColumn formats a top column row with bold caption and normal value
func renderTopColumn(caption, value string) string {
	if strings.TrimSpace(caption) != "" {
		caption += ": "
	} else {
		caption += "  "
	}

	return bold.Render(caption) + value
}

func renderProgress(progress int32) string {
	var stl lipgloss.Style

	switch {
	case progress < 20:
		stl = red
	case progress < 40:
		stl = darkorange
	case progress < 60:
		stl = orange
	case progress < 80:
		stl = gold
	case progress > 81 && progress < 100:
		stl = yellow
	case progress == 100:
		stl = green
	}

	return stl.Render(fmt.Sprintf("%d%%", progress))
}
