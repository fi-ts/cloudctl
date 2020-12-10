package output

import (
	"fmt"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// VolumeTablePrinter prints volumes in a table
	VolumeTablePrinter struct {
		TablePrinter
	}
)

// Print an ip as table
func (p VolumeTablePrinter) Print(data []*models.V1VolumeResponse) {
	p.wideHeader = []string{"ID", "Size", "Replicas", "Project", "Partition", "Nodes"}
	p.shortHeader = p.wideHeader

	for _, vol := range data {
		volumeID := ""
		if vol.VolumeID != nil {
			volumeID = *vol.VolumeID
		}
		size := ""
		if vol.Size != nil {
			size = fmt.Sprintf("%d", *vol.Size)
		}
		replica := ""
		if vol.ReplicaCount != nil {
			replica = fmt.Sprintf("%d", *vol.ReplicaCount)
		}
		partition := ""
		if vol.PartitionID != nil {
			partition = *vol.PartitionID
		}
		project := ""
		if vol.ProjectID != nil {
			project = *vol.ProjectID
		}

		wide := []string{volumeID, size, replica, project, partition, strings.Join(vol.ConnectedHosts, "\n")}
		short := wide

		p.addWideData(wide, vol)
		p.addShortData(short, vol)
	}
	p.render()
}
