package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// VolumeTablePrinter prints volumes in a table
	VolumeTablePrinter struct {
		TablePrinter
	}
)

// Print an ip as table
func (p VolumeTablePrinter) Print(data []*models.V1VolumeResponse) {
	p.wideHeader = []string{"ID", "Size", "Usage", "Replicas", "StorageClass", "Project", "Partition", "Nodes"}
	p.shortHeader = p.wideHeader

	for _, vol := range data {
		volumeID := ""
		if vol.VolumeID != nil {
			volumeID = *vol.VolumeID
		}
		size := ""
		if vol.Size != nil {
			size = fmt.Sprintf("%s", humanize.IBytes(uint64(*vol.Size)))
		}
		usage := ""
		if vol.Statistics != nil && vol.Statistics.LogicalUsedStorage != nil {
			usage = fmt.Sprintf("%s", humanize.IBytes(uint64(*vol.Statistics.LogicalUsedStorage)))
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

		nodes := ConnectedHosts(vol)

		wide := []string{volumeID, size, usage, replica, sc, project, partition, strings.Join(nodes, "\n")}
		short := wide

		p.addWideData(wide, vol)
		p.addShortData(short, vol)
	}
	p.render()
}

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
Name:            pvc-e86fe06d-7d5e-44e0-b83b-e6758caa2826
Labels:          <none>
Annotations:     pv.kubernetes.io/provisioned-by: csi.lightbitslabs.com
Finalizers:      [kubernetes.io/pv-protection external-attacher/csi-lightbitslabs-com]
StorageClass:    partition-storage-gold
Status:          Bound
Claim:           default/example-mt-pvc
Reclaim Policy:  Delete
Access Modes:    RWO
VolumeMode:      Filesystem
Capacity:        10Gi
Node Affinity:   <none>
Message:
Source:
    Type:              CSI (a Container Storage Interface (CSI) volume source)
    Driver:            csi.lightbitslabs.com
    FSType:            ext4
	VolumeHandle:      mgmt:10.131.44.1:443,10.131.44.2:443,10.131.44.3:443|nguid:9304ff62-2d56-4d67-9c3d-80585a6795df|proj:b5f26a3b-9a4d-48db-a6b3-d1dd4ac4abec|scheme:grpcs
	                                                                              <volume.UUID>                             <volume.ProjectName>
    ReadOnly:          false
    VolumeAttributes:      storage.kubernetes.io/csiProvisionerIdentity=1607936819649-8081-csi.lightbitslabs.com
Events:                <none>
*/
func PersistenVolume(v models.V1VolumeResponse, name, namespace string) error {
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
	js, err := json.Marshal(pv)
	if err != nil {
		return fmt.Errorf("unable to marshal to yaml:%v", err)
	}
	y, err := yaml.JSONToYAML(js)
	fmt.Printf("%s\n", string(y))
	return nil
}
