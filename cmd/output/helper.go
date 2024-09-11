package output

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
)

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

func sortIPs(v1ips []*models.ModelsV1IPResponse) []*models.ModelsV1IPResponse {

	v1ipmap := make(map[string]*models.ModelsV1IPResponse)
	var ips []string
	for _, v1ip := range v1ips {
		v1ipmap[*v1ip.Ipaddress] = v1ip
		ips = append(ips, *v1ip.Ipaddress)
	}

	realIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		realIPs = append(realIPs, net.ParseIP(ip))
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	var result []*models.ModelsV1IPResponse
	for _, ip := range realIPs {
		result = append(result, v1ipmap[ip.String()])
	}
	return result
}

func printStringSlice(s []string) {
	var dashed []string
	for _, elem := range s {
		dashed = append(dashed, "- "+elem)
	}
	fmt.Println(strings.Join(dashed, "\n"))
}
