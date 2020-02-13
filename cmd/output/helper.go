package output

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
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

// strValue returns the value of a string pointer of not nil, otherwise empty string
func strValue(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return ""
}

// FIXME write a test
func truncate(input, elipsis string, maxlength int) string {
	il := len(input)
	el := len(elipsis)
	if il <= maxlength {
		return input
	}
	if maxlength <= el {
		return input[:maxlength]
	}
	startlength := ((maxlength - el) / 2) - el/2

	output := input[:startlength] + elipsis
	missing := maxlength - len(output)
	output = output + input[il-missing:]
	return output
}
