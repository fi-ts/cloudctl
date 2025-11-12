package output

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"
)

type (
	// VolumeTablePrinter prints volumes in a table
	VolumeTablePrinter struct {
		tablePrinter
	}
	VolumeClusterInfoTablePrinter struct {
		tablePrinter
	}
	SnapshotTablePrinter struct {
		tablePrinter
	}
	QoSPolicyTablePrinter struct {
		tablePrinter
	}
)

// Print a volume as table
func (p VolumeTablePrinter) Print(data []*models.V1VolumeResponse) {
	p.shortHeader = []string{"ID", "Name", "Size", "Usage", "Replicas", "QoS", "Project", "Tenant", "Partition"}
	p.wideHeader = append(p.shortHeader, "Nodes")
	p.Order(data)

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
			size = humanize.IBytes(uint64(*vol.Size)) // nolint:gosec
		}
		usage := ""
		if vol.Statistics != nil && vol.Statistics.LogicalUsedStorage != nil {
			usage = humanize.IBytes(uint64(*vol.Statistics.LogicalUsedStorage)) // nolint:gosec
		}
		replica := ""
		if vol.ReplicaCount != nil {
			replica = fmt.Sprintf("%d", *vol.ReplicaCount)
		}
		qos := ""
		if vol.QosPolicyName != nil {
			qos = *vol.QosPolicyName
		} else if vol.QosPolicyUUID != nil {
			qos = *vol.QosPolicyUUID
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

		short := []string{volumeID, name, size, usage, replica, qos, project, tenant, partition}
		wide := append(short, strings.Join(nodes, "\n"))

		p.addWideData(wide, vol)
		p.addShortData(short, vol)
	}
	p.render()
}

// Print an snapshot as table
func (p SnapshotTablePrinter) Print(data []*models.V1SnapshotResponse) {
	p.shortHeader = []string{"ID", "Name", "SourceVolumeID", "SourceVolumeName", "Size", "Project", "Tenant", "Partition"}
	p.wideHeader = append(p.shortHeader, "Nodes")
	p.Order(data)

	for _, snap := range data {
		snapshotID := ""
		if snap.SnapshotID != nil {
			snapshotID = *snap.SnapshotID
		}
		name := ""
		if snap.Name != nil {
			name = *snap.Name
		}
		size := ""
		if snap.Size != nil {
			size = humanize.IBytes(uint64(*snap.Size)) // nolint:gosec
		}
		partition := ""
		if snap.PartitionID != nil {
			partition = *snap.PartitionID
		}
		project := ""
		if snap.ProjectID != nil {
			project = *snap.ProjectID
		}
		tenant := ""
		if snap.TenantID != nil {
			tenant = *snap.TenantID
		}
		sourceID := ""
		if snap.SourceVolumeID != nil {
			sourceID = *snap.SourceVolumeID
		}
		sourceName := ""
		if snap.SourceVolumeName != nil {
			sourceName = *snap.SourceVolumeName
		}

		short := []string{snapshotID, name, sourceID, sourceName, size, project, tenant, partition}
		wide := short

		p.addWideData(wide, snap)
		p.addShortData(short, snap)
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
func VolumeManifest(v models.V1VolumeResponse, name, namespace, sc string) error {
	filesystem := corev1.PersistentVolumeFilesystem
	pv := corev1.PersistentVolume{
		TypeMeta:   v1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeMode:       &filesystem,
			StorageClassName: sc,
			// FIXME add Capacity once figured out
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
		fmt.Printf("# be cautious! at the time being your volume:%s is still attached to worker node:%s, you can not mount it twice\n", *v.VolumeID, strings.Join(nodes, ","))
	}

	helper.MustPrintKubernetesResource(pv)
	return nil
}

func VolumeEncryptionSecretManifest(namespace, passphrase string) error {
	secret := corev1.Secret{
		TypeMeta: v1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: v1.ObjectMeta{
			Name:      "storage-encryption-key",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"host-encryption-passphrase": passphrase,
		},
	}
	y, err := k8syaml.Marshal(secret)
	if err != nil {
		return err
	}
	fmt.Println(`# Sample secret to be used in conjunction with the partition-gold-encrypted StorageClass.
# Place this secret, after modifying namespace and the actual secret value, in the same namespace as the pvc.
#
# IMPORTANT
# Remember to make a safe copy of this secret at a secure location, once lost all your data will be lost as well.`)
	fmt.Println(string(y))
	return nil
}

func (p VolumeClusterInfoTablePrinter) Print(data []*models.V1StorageClusterInfo) {
	p.wideHeader = []string{"Partition", "Version", "Health", "Nodes NA", "Volumes D/NA/RO", "Physical Installed/Managed", "Physical Effective/Free/Used", "Logical Total/Used", "Estimated Total/Free", "Compression"}
	p.shortHeader = p.wideHeader

	for _, info := range data {

		if info == nil || info.Statistics == nil {
			continue
		}

		partition := pointer.SafeDeref(info.Partition)
		health := pointer.SafeDeref(info.Health.State)
		numdegradedvolumes := pointer.SafeDeref(info.Health.NumDegradedVolumes)
		numnotavailablevolumes := pointer.SafeDeref(info.Health.NumNotAvailableVolumes)
		numreadonlyvolumes := pointer.SafeDeref(info.Health.NumReadOnlyVolumes)
		numinactivenodes := pointer.SafeDeref(info.Health.NumInactiveNodes)

		compressionratio := ""
		if info.Statistics != nil && info.Statistics.CompressionRatio != nil {
			ratio := *info.Statistics.CompressionRatio
			compressionratio = fmt.Sprintf("%d%%", int(100.0*(1-ratio)))
		}
		effectivephysicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.EffectivePhysicalStorage))
		freephysicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.FreePhysicalStorage))
		physicalusedstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.PhysicalUsedStorage))

		estimatedfreelogicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.EstimatedFreeLogicalStorage))
		estimatedlogicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.EstimatedLogicalStorage))
		logicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.LogicalStorage))
		logicalusedstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.LogicalUsedStorage))
		installedphysicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.InstalledPhysicalStorage))
		managedphysicalstorage := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.ManagedPhysicalStorage))
		// physicalusedstorageincludingparity := helper.HumanizeSize(pointer.SafeDeref(info.Statistics.PhysicalUsedStorageIncludingParity))

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

// Print a QoS Policy as table
func (p QoSPolicyTablePrinter) Print(data []*models.V1QoSPolicyResponse) {
	p.shortHeader = []string{"ID", "Name", "Partition", "Description", "State", "Read", "Write"}
	p.wideHeader = p.shortHeader

	for _, qos := range data {
		id := ""
		if qos.QoSPolicyID != nil {
			id = *qos.QoSPolicyID
		}
		name := ""
		if qos.Name != nil {
			name = *qos.Name
		}
		partition := ""
		if qos.Partition != nil {
			partition = *qos.Partition
		}
		description := ""
		longDescription := ""
		if qos.Description != nil {
			longDescription = *qos.Description
			description = genericcli.TruncateEnd(*qos.Description, 40)
		}
		state := ""
		if qos.State != nil {
			state = *qos.State
		}
		read := ""
		write := ""
		if qos.Limit != nil {
			if l := qos.Limit.Bandwidth; l != nil {
				if l.Read != nil {
					read = fmt.Sprintf("%d MB/s", *l.Read)
				}
				if l.Write != nil {
					write = fmt.Sprintf("%d MB/s", *l.Write)
				}
			}
			if l := qos.Limit.IOPS; l != nil {
				if l.Read != nil {
					read = fmt.Sprintf("%d IOPS", *l.Read)
				}
				if l.Write != nil {
					write = fmt.Sprintf("%d IOPS", *l.Write)
				}
			}
			if l := qos.Limit.IOPSPerGB; l != nil {
				if l.Read != nil {
					read = fmt.Sprintf("%d IOPS/GB", *l.Read)
				}
				if l.Write != nil {
					write = fmt.Sprintf("%d IOPS/GB", *l.Write)
				}
			}
		}

		short := []string{id, name, partition, description, state, read, write}
		wide := []string{id, name, partition, longDescription, state, read, write}

		p.addWideData(wide, wide)
		p.addShortData(short, qos)
	}
	p.render()
}
