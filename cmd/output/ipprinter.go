package output

import (
	"strings"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/helper"
	metalgo "github.com/metal-pod/metal-go"
)

type (
	// IPTablePrinter prints ips in a table
	IPTablePrinter struct {
		TablePrinter
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
			if strings.HasPrefix(t, metalgo.TagMachinePrefix+"=") {
				shortTags = append(shortTags, "machine:"+parts[1])
			} else if strings.HasPrefix(t, metalgo.TagServicePrefix+"=") {
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
