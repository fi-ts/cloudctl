package sorters

import (
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func S3Sorter() *multisort.Sorter[*models.V1S3CredentialsResponse] {
	return multisort.New(multisort.FieldMap[*models.V1S3CredentialsResponse]{
		"id": func(a, b *models.V1S3CredentialsResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.ID), pointer.SafeDeref(b.ID), descending)
		},
		"name": func(a, b *models.V1S3CredentialsResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Name), pointer.SafeDeref(b.Name), descending)
		},
		"tenant": func(a, b *models.V1S3CredentialsResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Tenant), pointer.SafeDeref(b.Tenant), descending)
		},
		"project": func(a, b *models.V1S3CredentialsResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Project), pointer.SafeDeref(b.Project), descending)
		},
		"partition": func(a, b *models.V1S3CredentialsResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Partition), pointer.SafeDeref(b.Partition), descending)
		},
	}, multisort.Keys{{ID: "tenant"}, {ID: "partition"}, {ID: "project"}, {ID: "id"}})
}
