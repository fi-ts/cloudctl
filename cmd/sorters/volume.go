package sorters

import (
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func VolumeSorter() *multisort.Sorter[*models.V1VolumeResponse] {
	return multisort.New(multisort.FieldMap[*models.V1VolumeResponse]{
		"id": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.VolumeID), p.SafeDeref(b.VolumeID), descending)
		},
		"name": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.VolumeName), p.SafeDeref(b.VolumeName), descending)
		},
		"project": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.ProjectID), p.SafeDeref(b.ProjectID), descending)
		},
		"tenant": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.TenantID), p.SafeDeref(b.TenantID), descending)
		},
		"partition": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.PartitionID), p.SafeDeref(b.PartitionID), descending)
		},
		"usage": func(a, b *models.V1VolumeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(pointer.SafeDeref(a.Statistics).LogicalUsedStorage), pointer.SafeDeref(pointer.SafeDeref(b.Statistics).LogicalUsedStorage), descending)
		},
	}, multisort.Keys{{ID: "tenant"}, {ID: "project"}, {ID: "partition"}, {ID: "id"}})
}

func SnapshotSorter() *multisort.Sorter[*models.V1SnapshotResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SnapshotResponse]{
		"id": func(a, b *models.V1SnapshotResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.SnapshotID), p.SafeDeref(b.SnapshotID), descending)
		},
		"volume_name": func(a, b *models.V1SnapshotResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.SourceVolumeName), p.SafeDeref(b.SourceVolumeName), descending)
		},
		"tenant": func(a, b *models.V1SnapshotResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.TenantID), p.SafeDeref(b.TenantID), descending)
		},
		"partition": func(a, b *models.V1SnapshotResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.PartitionID), p.SafeDeref(b.PartitionID), descending)
		},
	}, multisort.Keys{{ID: "tenant"}, {ID: "partition"}, {ID: "id"}})
}
