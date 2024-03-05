package output

import (
	"sort"

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

	keys := make([]string, 0, len(services))
	for k := range services {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		s := services[name]

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

		i := 0
		for sname, sresult := range s.Services {
			prefix := "├"
			if i == len(s.Services)-1 {
				prefix = "└"
			}
			prefix += "─╴"

			status := "unknown"
			if sresult.Status != nil && *sresult.Status != "" {
				status = *sresult.Status
			}
			msg := ""
			if sresult.Message != nil && *sresult.Message != "" {
				msg = *s.Message
			}

			wide := []string{prefix + sname, status, msg}
			p.addWideData(wide, s)
			p.addShortData(wide, s)
			i++
		}
	}

	p.render()
}
