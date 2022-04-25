package output

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// VolumeTablePrinter prints volumes in a table
	VolumeTablePrinter struct {
		tablePrinter
	}
	VolumeClusterInfoTablePrinter struct {
		tablePrinter
	}
)

// Print an volume as table
func (p VolumeTablePrinter) Print(data []*models.V1VolumeResponse) {
	p.shortHeader = []string{"ID", "Name", "Size", "Usage", "Replicas", "StorageClass", "Project", "Tenant", "Partition"}
	p.wideHeader = append(p.shortHeader, "Nodes")

	for _, vol := range data {
		volumeID := ""
		if vol.VolumeID != nil {
			volumeID = *vol.VolumeID
		}
		name := ""
		if vol.VolumeName != nil {
			name = *vol.VolumeName
		}
		size := ""
		if vol.Size != nil {
			size = humanize.IBytes(uint64(*vol.Size))
		}
		usage := ""
		if vol.Statistics != nil && vol.Statistics.PhysicalUsedStorage != nil {
			usage = humanize.IBytes(uint64(*vol.Statistics.PhysicalUsedStorage))
		}
		replica := ""
		if vol.ReplicaCount != nil {
			replica = fmt.Sprintf("%d", *vol.ReplicaCount)
		}
		sc := ""
		if vol.StorageClass != nil {
			sc = *vol.StorageClass
		}
		partition := ""
		if vol.PartitionID != nil {
			partition = *vol.PartitionID
		}
		project := ""
		if vol.ProjectID != nil {
			project = *vol.ProjectID
		}
		tenant := ""
		if vol.TenantID != nil {
			tenant = *vol.TenantID
		}

		nodes := ConnectedHosts(vol)

		short := []string{volumeID, name, size, usage, replica, sc, project, tenant, partition}
		wide := append(short, strings.Join(nodes, "\n"))

		p.addWideData(wide, vol)
		p.addShortData(short, vol)
	}
	p.render()
}

// ConnectedHosts returns the worker nodes without internal prefixes and suffixes
func ConnectedHosts(vol *models.V1VolumeResponse) []string {
	nodes := []string{}
	for _, n := range vol.ConnectedHosts {
		// nqn.2019-09.com.lightbitslabs:host:shoot--pddhz9--duros-tst9-group-0-6b7bb-2cnvs.node
		parts := strings.Split(n, ":host:")
		if len(parts) >= 1 {
			node := strings.TrimSuffix(parts[1], ".node")
			nodes = append(nodes, node)
		}
	}
	return nodes
}

/*
VolumeManifest create a manifest for static PV like so

---
apiVersion: v1
kind: PersistentVolume
metadata:
  annotations:
    pv.kubernetes.io/provisioned-by: csi.lightbitslabs.com
  name: pvc-7e3b4b43-0061-46f0-a125-e0c1a0b2a4fb
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 20Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: example-mt-pvc-2
    namespace: default
    resourceVersion: "13088"
    uid: 7e3b4b43-0061-46f0-a125-e0c1a0b2a4fb
  csi:
    controllerExpandSecretRef:
      name: lb-csi-creds
      namespace: kube-system
    controllerPublishSecretRef:
      name: lb-csi-creds
      namespace: kube-system
    driver: csi.lightbitslabs.com
    fsType: ext4
    nodePublishSecretRef:
      name: lb-csi-creds
      namespace: kube-system
    nodeStageSecretRef:
      name: lb-csi-creds
      namespace: kube-system
    volumeAttributes:
      storage.kubernetes.io/csiProvisionerIdentity: 1608281061542-8081-csi.lightbitslabs.com
    volumeHandle: mgmt:10.131.44.1:443,10.131.44.2:443,10.131.44.3:443|nguid:d798ec5a-84b3-4909-9938-785ebd3bbadb|proj:24235d11-deb9-46e3-9101-f906c78b10b6|scheme:grpcs
  persistentVolumeReclaimPolicy: Delete
  storageClassName: partition-silver
*/
func VolumeManifest(v models.V1VolumeResponse, name, namespace string) error {
	filesystem := corev1.PersistentVolumeFilesystem
	pv := corev1.PersistentVolume{
		TypeMeta:   v1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeMode:  &filesystem,
			// FIXME add Capacity once figured out
			StorageClassName: *v.StorageClass,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver:       "csi.lightbitslabs.com",
					FSType:       "ext4",
					ReadOnly:     false,
					VolumeHandle: *v.VolumeHandle,
				},
			},
		},
	}

	if len(v.ConnectedHosts) > 0 {
		nodes := ConnectedHosts(&v)
		fmt.Printf("# be cautios! at the time being your volume:%s is still attached to worker node:%s, you can not mount it twice\n", *v.VolumeID, strings.Join(nodes, ","))
	}

	helper.MustPrintKubernetesResource(pv)
	return nil
}

func (p VolumeClusterInfoTablePrinter) Print(data []*models.V1StorageClusterInfo) {
	p.wideHeader = []string{"Partition", "Version", "Health", "Nodes NA", "Volumes D/NA/RO", "Physical Installed/Managed", "Physical Effective/Free/Used", "Logical Total/Used", "Estimated Total/Free", "Compression"}
	p.shortHeader = p.wideHeader

	for _, info := range data {

		if info == nil || info.Statistics == nil {
			continue
		}

		partition := strValue(info.Partition)
		health := strValue(info.Health.State)
		numdegradedvolumes := int64Value(info.Health.NumDegradedVolumes)
		numnotavailablevolumes := int64Value(info.Health.NumNotAvailableVolumes)
		numreadonlyvolumes := int64Value(info.Health.NumReadOnlyVolumes)
		numinactivenodes := int64Value(info.Health.NumInactiveNodes)

		compressionratio := ""
		if info.Statistics != nil && info.Statistics.CompressionRatio != nil {
			ratio := *info.Statistics.CompressionRatio
			compressionratio = fmt.Sprintf("%d%%", int(100.0*(1-ratio)))
		}
		effectivephysicalstorage := helper.HumanizeSize(int64Value(info.Statistics.EffectivePhysicalStorage))
		freephysicalstorage := helper.HumanizeSize(int64Value(info.Statistics.FreePhysicalStorage))
		physicalusedstorage := helper.HumanizeSize(int64Value(info.Statistics.PhysicalUsedStorage))

		estimatedfreelogicalstorage := helper.HumanizeSize(int64Value(info.Statistics.EstimatedFreeLogicalStorage))
		estimatedlogicalstorage := helper.HumanizeSize(int64Value(info.Statistics.EstimatedLogicalStorage))
		logicalstorage := helper.HumanizeSize(int64Value(info.Statistics.LogicalStorage))
		logicalusedstorage := helper.HumanizeSize(int64Value(info.Statistics.LogicalUsedStorage))
		installedphysicalstorage := helper.HumanizeSize(int64Value(info.Statistics.InstalledPhysicalStorage))
		managedphysicalstorage := helper.HumanizeSize(int64Value(info.Statistics.ManagedPhysicalStorage))
		// physicalusedstorageincludingparity := helper.HumanizeSize(int64Value(info.Statistics.PhysicalUsedStorageIncludingParity))

		version := "n/a"
		if info.MinVersionInCluster != nil {
			version = *info.MinVersionInCluster
		}
		wide := []string{
			partition,
			version,
			health,
			fmt.Sprintf("%d", numinactivenodes),
			fmt.Sprintf("%d/%d/%d", numdegradedvolumes, numnotavailablevolumes, numreadonlyvolumes),
			installedphysicalstorage + "/" + managedphysicalstorage,
			effectivephysicalstorage + "/" + freephysicalstorage + "/" + physicalusedstorage,
			logicalstorage + "/" + logicalusedstorage,
			estimatedlogicalstorage + "/" + estimatedfreelogicalstorage,
			compressionratio,
		}
		short := wide

		p.addWideData(wide, info)
		p.addShortData(short, info)
	}
	p.render()
}
