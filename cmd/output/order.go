package output

import (
	"bytes"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/tag"
)

// Order cluster
func (s ShootTablePrinter) Order(data []*models.V1ClusterResponse) {
	cols := strings.Split(s.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			tenantA := A.Shoot.Metadata.Annotations[tag.ClusterTenant]
			tenantB := B.Shoot.Metadata.Annotations[tag.ClusterTenant]
			projectA := A.Shoot.Metadata.Annotations[tag.ClusterProject]
			projectB := B.Shoot.Metadata.Annotations[tag.ClusterProject]
			nameA := A.Shoot.Metadata.Name
			nameB := B.Shoot.Metadata.Name
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "tenant":
					if tenantA == "" {
						return true
					}
					if tenantB == "" {
						return false
					}
					if tenantA < tenantB {
						return true
					}
					if tenantA != tenantB {
						return false
					}
				case "project":
					if projectA == "" {
						return true
					}
					if projectB == "" {
						return false
					}
					if projectA < projectB {
						return true
					}
					if projectA != projectB {
						return false
					}
				case "name":
					if nameA == "" {
						return true
					}
					if nameB == "" {
						return false
					}
					if nameA < nameB {
						return true
					}
					if nameA != nameB {
						return false
					}
				}
			}
			return false
		})
	}
}

// Order Project
func (s ProjectTablePrinter) Order(data []*models.V1Project) {
	cols := strings.Split(s.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "tenant":
					if A.TenantID == "" {
						return true
					}
					if B.TenantID == "" {
						return false
					}
					if A.TenantID < B.TenantID {
						return true
					}
					if A.TenantID != B.TenantID {
						return false
					}
				case "project":
					if A.Name == "" {
						return true
					}
					if B.Name == "" {
						return false
					}
					if A.Name < B.Name {
						return true
					}
					if A.Name != B.Name {
						return false
					}
				}
			}

			return false
		})
	}
}

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

// Order ipUsage
func (s *IPBillingTablePrinter) Order(data []*models.V1IPUsage) {
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
					if A.Projectid == nil {
						return true
					}
					if B.Projectid == nil {
						return false
					}
					if *A.Projectid < *B.Projectid {
						return true
					}
					if *A.Projectid != *B.Projectid {
						return false
					}
				case "ip":
					if A.IP == nil {
						return true
					}
					if B.IP == nil {
						return false
					}
					ipA := net.ParseIP(*A.IP)
					if ipA == nil {
						return true
					}
					ipB := net.ParseIP(*B.IP)
					if ipB == nil {
						return false
					}
					if bytes.Compare(ipA, ipB) < 0 {
						return true
					}
					if !ipA.Equal(ipB) {
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

// Order volumeUsage
func (s *NetworkTrafficBillingTablePrinter) Order(data []*models.V1NetworkUsage) {
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
				case "device":
					if A.Device == nil {
						return true
					}
					if B.Device == nil {
						return false
					}
					if *A.Device < *B.Device {
						return true
					}
					if *A.Device != *B.Device {
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

// Order s3Usage
func (s *S3BillingTablePrinter) Order(data []*models.V1S3Usage) {
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
				case "user":
					if A.User == nil {
						return true
					}
					if B.User == nil {
						return false
					}
					if *A.User < *B.User {
						return true
					}
					if *A.User != *B.User {
						return false
					}
				case "bucket":
					if A.Bucketname == nil {
						return true
					}
					if B.Bucketname == nil {
						return false
					}
					if *A.Bucketname < *B.Bucketname {
						return true
					}
					if *A.Bucketname != *B.Bucketname {
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

// Order volumeUsage
func (s *VolumeBillingTablePrinter) Order(data []*models.V1VolumeUsage) {
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

// Order machines
func (m MachineTablePrinter) Order(data []*models.ModelsV1MachineResponse) {
	cols := strings.Split(m.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "features":
					a := A.Allocation.Image.Features[0]
					b := B.Allocation.Image.Features[0]
					if a < b {
						return true
					}
					if a != b {
						return false
					}
				case "hostname":
					if A.Allocation.Hostname == nil {
						return true
					}
					if B.Allocation.Hostname == nil {
						return false
					}
					if *A.Allocation.Hostname < *B.Allocation.Hostname {
						return true
					}
					if *A.Allocation.Hostname != *B.Allocation.Hostname {
						return false
					}
				}
			}
			return false
		})
	}
}

// Order s3 partitions
func (m S3PartitionTablePrinter) Order(data []*models.V1S3PartitionResponse) {
	cols := strings.Split(m.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "id":
					if A.ID == nil {
						return true
					}
					if B.ID == nil {
						return false
					}
					if *A.ID < *B.ID {
						return true
					}
					if *A.ID != *B.ID {
						return false
					}
				}
			}
			return false
		})
	}
}
