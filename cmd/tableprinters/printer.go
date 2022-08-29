package tableprinters

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

const (
	circle = "●"
)

type TablePrinter struct {
	t *printers.TablePrinter
}

func New() *TablePrinter {
	return &TablePrinter{}
}

func (t *TablePrinter) SetPrinter(printer *printers.TablePrinter) {
	t.t = printer
}

func (t *TablePrinter) ToHeaderAndRows(data any, wide bool) ([]string, [][]string, error) {
	switch d := data.(type) {
	case *models.V1S3CredentialsResponse:
		return t.S3Table(pointer.WrapInSlice(d), wide)
	case []*models.V1S3CredentialsResponse:
		return t.S3Table(d, wide)
	case *models.V1S3PartitionResponse:
		return t.S3PartitionsTable(pointer.WrapInSlice(d), wide)
	case []*models.V1S3PartitionResponse:
		return t.S3PartitionsTable(d, wide)
	case *models.V1ProjectResponse:
		return t.ProjectTable(pointer.WrapInSlice(d), wide)
	case []*models.V1ProjectResponse:
		return t.ProjectTable(d, wide)
	case *models.V1TenantResponse:
		return t.TenantTable(pointer.WrapInSlice(d), wide)
	case []*models.V1TenantResponse:
		return t.TenantTable(d, wide)
	case *models.V1SnapshotResponse:
		return t.SnapshotTable(pointer.WrapInSlice(d), wide)
	case []*models.V1SnapshotResponse:
		return t.SnapshotTable(d, wide)
	case *models.V1VolumeResponse:
		return t.VolumeTable(pointer.WrapInSlice(d), wide)
	case []*models.V1VolumeResponse:
		return t.VolumeTable(d, wide)
	case *models.V1StorageClusterInfo:
		return t.VolumeClusterInfoTable(pointer.WrapInSlice(d), wide)
	case []*models.V1StorageClusterInfo:
		return t.VolumeClusterInfoTable(d, wide)
	default:
		return nil, nil, fmt.Errorf("unknown table printer for type: %T", d)
	}
}
