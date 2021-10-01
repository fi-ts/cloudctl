package output

import (
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/metal-stack/metal-lib/pkg/tag"
)

type (
	// IPTablePrinter prints ips in a table
	IPTablePrinter struct {
		tablePrinter
	}
)

// Print an ip as table
func (p IPTablePrinter) Print(data []*models.ModelsV1IPResponse) {
	data = sortIPs(data)
	p.wideHeader = []string{"IP", "Type", "Name", "Description", "Network", "Project", "Tags"}
	p.shortHeader = []string{"IP", "Type", "Name", "Network", "Project", "Tags"}

	for _, ip := range data {
		networkid := ""
		if ip.Networkid != nil {
			networkid = *ip.Networkid
		}
		projectid := ""
		if ip.Projectid != nil {
			projectid = *ip.Projectid
		}

		truncatedName := helper.Truncate(ip.Name, "...", 30)

		var shortTags []string
		for _, t := range ip.Tags {
			parts := strings.Split(t, "=")
			if strings.HasPrefix(t, tag.MachineID+"=") {
				shortTags = append(shortTags, "machine:"+parts[1])
			} else if strings.HasPrefix(t, tag.ClusterServiceFQN+"=") {
				shortTags = append(shortTags, "service:"+parts[1])
			} else {
				shortTags = append(shortTags, t)
			}
		}

		wide := []string{*ip.Ipaddress, *ip.Type, ip.Name, ip.Description, networkid, projectid, strings.Join(ip.Tags, "\n")}
		short := []string{*ip.Ipaddress, *ip.Type, truncatedName, networkid, projectid, strings.Join(shortTags, "\n")}

		p.addWideData(wide, ip)
		p.addShortData(short, ip)
	}
	p.render()
}
