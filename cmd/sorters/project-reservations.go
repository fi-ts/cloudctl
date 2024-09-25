package sorters

import (
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func MachineReservationsSorter() *multisort.Sorter[*models.V1MachineReservationResponse] {
	return multisort.New(multisort.FieldMap[*models.V1MachineReservationResponse]{
		"id": func(a, b *models.V1MachineReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
		},
		"tenant": func(a, b *models.V1MachineReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Tenant), p.SafeDeref(b.Tenant), descending)
		},
		"project": func(a, b *models.V1MachineReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Projectid), p.SafeDeref(b.Projectid), descending)
		},
		"size": func(a, b *models.V1MachineReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Sizeid), p.SafeDeref(b.Sizeid), descending)
		},
		"amount": func(a, b *models.V1MachineReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Amount), p.SafeDeref(b.Amount), descending)
		},
	}, multisort.Keys{{ID: "tenant"}, {ID: "project"}, {ID: "size"}, {ID: "id"}})
}

func MachineReservationsUsageSorter() *multisort.Sorter[*models.V1MachineReservationUsageResponse] {
	return multisort.New(multisort.FieldMap[*models.V1MachineReservationUsageResponse]{
		"id": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
		},
		"tenant": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Tenant), p.SafeDeref(b.Tenant), descending)
		},
		"project": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Projectid), p.SafeDeref(b.Projectid), descending)
		},
		"size": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Sizeid), p.SafeDeref(b.Sizeid), descending)
		},
		"partition": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Partitionid), p.SafeDeref(b.Partitionid), descending)
		},
		"reservations": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Reservations), p.SafeDeref(b.Reservations), descending)
		},
		"unused-reservations": func(a, b *models.V1MachineReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Usedreservations), p.SafeDeref(b.Usedreservations), descending)
		},
	}, multisort.Keys{{ID: "tenant"}, {ID: "project"}, {ID: "partition"}, {ID: "size"}, {ID: "id"}})
}
