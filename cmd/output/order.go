package output

import (
	"sort"
	"strconv"
	"strings"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
)

// Order clusterUsage
func (s *ClusterBillingTablePrinter) Order(data []*models.V1ClusterUsage) {
	cols := strings.Split(s.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "tenant":
					if A.Tenant == nil {
						return true
					}
					if B.Tenant == nil {
						return false
					}
					if *A.Tenant < *B.Tenant {
						return true
					}
					if *A.Tenant != *B.Tenant {
						return false
					}
				case "project":
					if A.Projectname == nil {
						return true
					}
					if B.Projectname == nil {
						return false
					}
					if *A.Projectname < *B.Projectname {
						return true
					}
					if *A.Projectname != *B.Projectname {
						return false
					}
				case "partition":
					if A.Partition == nil {
						return true
					}
					if B.Partition == nil {
						return false
					}
					if *A.Partition < *B.Partition {
						return true
					}
					if *A.Partition != *B.Partition {
						return false
					}
				case "cluster":
					if A.Clustername == nil {
						return true
					}
					if B.Clustername == nil {
						return false
					}
					if *A.Clustername < *B.Clustername {
						return true
					}
					if *A.Clustername != *B.Clustername {
						return false
					}
				case "lifetime":
					if A.Lifetime == nil {
						return true
					}
					if B.Lifetime == nil {
						return false
					}
					aseconds := int64(*A.Lifetime)
					bseconds := int64(*B.Lifetime)
					if aseconds < bseconds {
						return true
					}
					if aseconds != bseconds {
						return false
					}
				}
			}

			return false
		})
	}
}

// Order containerUsage
func (s *ContainerBillingTablePrinter) Order(data []*models.V1ContainerUsage) {
	cols := strings.Split(s.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "tenant":
					if A.Tenant == nil {
						return true
					}
					if B.Tenant == nil {
						return false
					}
					if *A.Tenant < *B.Tenant {
						return true
					}
					if *A.Tenant != *B.Tenant {
						return false
					}
				case "project":
					if A.Projectname == nil {
						return true
					}
					if B.Projectname == nil {
						return false
					}
					if *A.Projectname < *B.Projectname {
						return true
					}
					if *A.Projectname != *B.Projectname {
						return false
					}
				case "partition":
					if A.Partition == nil {
						return true
					}
					if B.Partition == nil {
						return false
					}
					if *A.Partition < *B.Partition {
						return true
					}
					if *A.Partition != *B.Partition {
						return false
					}
				case "cluster":
					if A.Clustername == nil {
						return true
					}
					if B.Clustername == nil {
						return false
					}
					if *A.Clustername < *B.Clustername {
						return true
					}
					if *A.Clustername != *B.Clustername {
						return false
					}
				case "namespace":
					if A.Namespace == nil {
						return true
					}
					if B.Namespace == nil {
						return false
					}
					if *A.Namespace < *B.Namespace {
						return true
					}
					if *A.Namespace != *B.Namespace {
						return false
					}
				case "pod":
					if A.Podname == nil {
						return true
					}
					if B.Podname == nil {
						return false
					}
					if *A.Podname < *B.Podname {
						return true
					}
					if *A.Podname != *B.Podname {
						return false
					}
				case "container":
					if A.Containername == nil {
						return true
					}
					if B.Containername == nil {
						return false
					}
					if *A.Containername < *B.Containername {
						return true
					}
					if *A.Containername != *B.Containername {
						return false
					}
				case "cpu":
					if A.Cpuseconds == nil {
						return true
					}
					if B.Cpuseconds == nil {
						return false
					}
					aseconds, err := strconv.ParseInt(*A.Cpuseconds, 10, 64)
					if err != nil {
						return true
					}
					bseconds, err := strconv.ParseInt(*B.Cpuseconds, 10, 64)
					if err != nil {
						return false
					}
					if aseconds < bseconds {
						return true
					}
					if aseconds != bseconds {
						return false
					}
				case "memory":
					if A.Memoryseconds == nil {
						return true
					}
					if B.Memoryseconds == nil {
						return false
					}
					aseconds, err := strconv.ParseInt(*A.Memoryseconds, 10, 64)
					if err != nil {
						return true
					}
					bseconds, err := strconv.ParseInt(*B.Memoryseconds, 10, 64)
					if err != nil {
						return false
					}
					if aseconds < bseconds {
						return true
					}
					if aseconds != bseconds {
						return false
					}
				}
			}

			return false
		})
	}
}
