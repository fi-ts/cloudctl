package output

import (
	"fmt"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
)

type (
	// AuditTablePrinter prints audit traces in a table
	AuditTablePrinter struct {
		tablePrinter
	}
)

// Print audit traces as table
func (p AuditTablePrinter) Print(data []*models.V1AuditResponse) {
	p.wideHeader = []string{"Time", "Request ID", "Component", "Detail", "Path", "Code", "User", "Tenant", "Body"}
	p.shortHeader = []string{"Time", "Request ID", "Component", "Detail", "Path", "Code", "User"}

	for _, trace := range data {
		var statusCode string
		if trace.StatusCode != 0 {
			statusCode = fmt.Sprintf("%d", trace.StatusCode)
		}
		row := []string{
			// using Local() is okay for user cli output
			time.Time(trace.Timestamp).Local().Format("2006-01-02 15:04:05 MST"), //nolint:gosmopolitan
			trace.Rqid,
			trace.Component,
			trace.Detail,
			trace.Path,
			statusCode,
			trace.User,
		}

		wide := append(row, trace.Tenant, genericcli.TruncateEnd(trace.Body, 40))

		p.addShortData(row, trace)
		p.addWideData(wide, trace)
	}
	p.render()
}
