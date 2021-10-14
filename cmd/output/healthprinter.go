package output

import (
	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// HealthTablePrinter
	HealthTablePrinter struct {
		tablePrinter
	}
)

func (p HealthTablePrinter) Print(health *models.RestHealthResponse) {
	p.wideHeader = []string{"Overall Status", "Message"}
	p.shortHeader = p.wideHeader

	status := "unknown"
	if health.Status != nil && *health.Status != "" {
		status = *health.Status
	}
	msg := ""
	if health.Message != nil && *health.Message != "" {
		msg = *health.Message
	}

	wide := []string{status, msg}
	p.addWideData(wide, health)
	p.addShortData(wide, health)

	p.render()
}

func (p HealthTablePrinter) PrintServices(services map[string]models.RestHealthResult) {
	p.wideHeader = []string{"Service", "Status", "Message"}
	p.shortHeader = p.wideHeader

	for name, s := range services {
		status := "unknown"
		if s.Status != nil && *s.Status != "" {
			status = *s.Status
		}
		msg := ""
		if s.Message != nil && *s.Message != "" {
			msg = *s.Message
		}

		wide := []string{name, status, msg}
		p.addWideData(wide, s)
		p.addShortData(wide, s)
	}

	p.render()
}
