package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// MachineTablePrinter print machines of Shoot Cluster in a Table
	MachineTablePrinter struct {
		tablePrinter
	}
)

const (
	nbr      = "\U00002007"
	skull    = "\U0001F480"
	question = "\U00002753"
	circle   = "●"
)

// Print a list of Machines in a table
func (m MachineTablePrinter) Print(data []*models.ModelsV1MachineResponse) {
	m.shortHeader = []string{"ID", "", "LAST EVENT", "WHEN", "STARTED", "AGE", "HOSTNAME", "IPs", "SIZE", "IMAGE", "PARTITION"}
	m.wideHeader = []string{"ID", "", "LAST EVENT", "WHEN", "STARTED", "AGE", "HOSTNAME", "IPs", "SIZE", "IMAGE", "PARTITION"}
	m.order = "features,hostname"
	m.Order(data)
	for _, machine := range data {
		machineID := *machine.ID

		alloc := machine.Allocation
		if alloc == nil {
			continue
		}
		// Needed ?
		// status := strValue(machine.Liveliness)
		var sizeID string
		if machine.Size != nil {
			sizeID = strValue(machine.Size.ID)
		}
		var partitionID string
		if machine.Partition != nil {
			partitionID = strValue(machine.Partition.ID)
		}
		hostname := strValue(alloc.Hostname)
		//truncatedHostname := truncate(hostname, "...", 30)

		var nwIPs []string
		for _, nw := range alloc.Networks {
			nwIPs = append(nwIPs, nw.Ips...)
		}
		ips := strings.Join(nwIPs, "\n")
		image := ""
		if alloc.Image != nil {
			image = strValue(alloc.Image.ID)
		}
		started := strValue(alloc.Created)
		age := ""
		format := "2006-01-02T15:04:05.999Z"
		created, err := time.Parse(format, *alloc.Created)
		if err != nil {
			fmt.Printf("unable to parse created time:%s", err)
			os.Exit(1)
		}
		if alloc.Created != nil && !created.IsZero() {
			started = created.Format(time.RFC3339)
			age = humanizeDuration(time.Since(created))
		}
		lastEvent := ""
		lastEventTime, err := time.Parse(format, machine.Events.LastEventTime)
		if err != nil {
			fmt.Printf("unable to parse lastevent time:%s", err)
			os.Exit(1)
		}
		when := ""
		if len(machine.Events.Log) > 0 {
			since := time.Since(lastEventTime)
			when = humanizeDuration(since)
			lastEvent = *machine.Events.Log[0].Event
		}
		status := strValue(machine.Liveliness)
		statusEmoji := ""
		switch status {
		case "Alive":
			statusEmoji = nbr
		case "Dead":
			statusEmoji = skull
		case "Unknown":
			statusEmoji = question
		default:
			statusEmoji = question
		}
		row := []string{machineID, statusEmoji, lastEvent, when, started, age, hostname, ips, sizeID, image, partitionID}
		m.addShortData(row, machine)
		m.addWideData(row, machine)
	}
	m.render()
}
