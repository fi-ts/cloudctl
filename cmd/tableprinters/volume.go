package tableprinters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) VolumeTable(data []*models.V1VolumeResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Tenant", "Project", "Partition", "Name", "Size", "Usage", "Replicas", "StorageClass"}
	if wide {
		header = append(header, "Nodes")
	}

	for _, vol := range data {
		row := []string{
			pointer.SafeDeref(vol.VolumeID),
			pointer.SafeDeref(vol.TenantID),
			pointer.SafeDeref(vol.ProjectID),
			pointer.SafeDeref(vol.PartitionID),
			pointer.SafeDeref(vol.VolumeName),
			humanize.IBytes(uint64(pointer.SafeDeref(vol.Size))),
			humanize.IBytes(uint64(pointer.SafeDeref(pointer.SafeDeref(vol.Statistics).LogicalUsedStorage))),
			strconv.FormatInt(pointer.SafeDeref(vol.ReplicaCount), 10),
			pointer.SafeDeref(vol.StorageClass),
		}
		if wide {
			nodes := ConnectedHosts(vol)
			row = append(row, strings.Join(nodes, "\n"))
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}

// ConnectedHosts returns the worker nodes without internal prefixes and suffixes
func ConnectedHosts(vol *models.V1VolumeResponse) []string {
	nodes := []string{}
	for _, n := range vol.ConnectedHosts {
		// nqn.2019-09.com.lightbitslabs:host:shoot--pddhz9--duros-tst9-group-0-6b7bb-2cnvs.node
		parts := strings.Split(n, ":host:")
		if len(parts) >= 1 {
			node := strings.TrimSuffix(parts[1], ".node")
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (t *TablePrinter) VolumeClusterInfoTable(data []*models.V1StorageClusterInfo, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"Partition", "Version", "Health", "Nodes NA", "Volumes D/NA/RO", "Physical Installed/Managed", "Physical Effective/Free/Used", "Logical Total/Used", "Estimated Total/Free", "Compression"}

	for _, info := range data {
		if info == nil {
			continue
		}

		health := pointer.SafeDeref(info.Health)
		statistics := pointer.SafeDeref(info.Statistics)

		effectivephysicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.EffectivePhysicalStorage))
		freephysicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.FreePhysicalStorage))
		physicalusedstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.PhysicalUsedStorage))

		estimatedfreelogicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.EstimatedFreeLogicalStorage))
		estimatedlogicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.EstimatedLogicalStorage))
		logicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.LogicalStorage))
		logicalusedstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.LogicalUsedStorage))
		installedphysicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.InstalledPhysicalStorage))
		managedphysicalstorage := helper.HumanizeSize(pointer.SafeDeref(statistics.ManagedPhysicalStorage))
		// physicalusedstorageincludingparity := helper.HumanizeSize(int64Value(info.Statistics.PhysicalUsedStorageIncludingParity))

		row := []string{
			pointer.SafeDeref(info.Partition),
			pointer.SafeDerefOrDefault(info.MinVersionInCluster, "n/a"),
			pointer.SafeDeref(info.Health.State),
			fmt.Sprintf("%d", pointer.SafeDeref(pointer.SafeDeref(info.Health).NumInactiveNodes)),
			fmt.Sprintf("%d/%d/%d", pointer.SafeDeref(health.NumDegradedVolumes), pointer.SafeDeref(health.NumNotAvailableVolumes), pointer.SafeDeref(health.NumReadOnlyVolumes)),
			installedphysicalstorage + "/" + managedphysicalstorage,
			effectivephysicalstorage + "/" + freephysicalstorage + "/" + physicalusedstorage,
			logicalstorage + "/" + logicalusedstorage,
			estimatedlogicalstorage + "/" + estimatedfreelogicalstorage,
			fmt.Sprintf("%d%%", int(100.0*(1-pointer.SafeDeref(statistics.CompressionRatio)))),
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}

func (t *TablePrinter) SnapshotTable(data []*models.V1SnapshotResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Tenant", "Partition", "Name", "SourceVolumeID", "SourceVolumeName", "Size"}

	for _, s := range data {
		row := []string{
			pointer.SafeDeref(s.SnapshotID),
			pointer.SafeDeref(s.TenantID),
			pointer.SafeDeref(s.PartitionID),
			pointer.SafeDeref(s.SnapshotID),
			pointer.SafeDeref(s.SourceVolumeID),
			pointer.SafeDeref(s.SourceVolumeName),
			humanize.IBytes(uint64(pointer.SafeDeref(s.Size))),
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}
