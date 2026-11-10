package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
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

// Machines returns a slice of cluster worker nodes
func (wd *watchData) Nodes() []*models.ModelsV1MachineResponse {
	ms := wd.shoot.Payload.Machines
	ms = append(ms, wd.shoot.Payload.Firewalls...)
	return ms
}

// Conditions returns a list of cluster status conditions
func (wd *watchData) Conditions() []*models.V1beta1Condition {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.Conditions
	}

	return nil
}

// LastOperation returns the last cluster operation
func (wd *watchData) LastOperation() *models.V1beta1LastOperation {
	if wd.shoot.Payload != nil && wd.shoot.Payload.Status != nil {
		return wd.shoot.Payload.Status.LastOperation
	}

	return nil
}

// LastErrors returns a slice of last cluster errors
func (wd *watchData) LastErrors() []*models.V1beta1LastError {
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

func (wd *watchData) progress() string {
	return fmt.Sprintf("%d%% [%s]",
		*wd.shoot.Payload.Status.LastOperation.Progress,
		*wd.shoot.Payload.Status.LastOperation.Type)
}

// watch is our tea.Model for bubbletea ELM
type watch struct {
	data          *watchData
	width, height int
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
		switch msg.String() {
		case "ctrl+c", "q":
			return w, tea.Quit
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
		// FIXME: add top row
		body.WriteString(w.data.id())
	}

	v := tea.NewView(body.String())
	v.AltScreen = true // fulscreen

	return v
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
