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

// Print an volume as table
func (p VolumeTablePrinter) Print(data []*models.V1VolumeResponse) {
	p.wideHeader = []string{"ID", "Size", "Usage", "Replicas", "StorageClass", "Project", "Tenant", "Partition", "Nodes"}
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
		tenant := ""
		if vol.TenantID != nil {
			tenant = *vol.TenantID
		}

		nodes := ConnectedHosts(vol)

		wide := []string{volumeID, size, usage, replica, sc, project, tenant, partition, strings.Join(nodes, "\n")}
		short := wide

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
PersistenVolume create a manifest for static PV like so

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

	if len(v.ConnectedHosts) > 0 {
		nodes := ConnectedHosts(&v)
		fmt.Printf("# be cautios! at the time being your volume:%s is still attached to worker node:%s, you can not mount it twice\n", *v.VolumeID, strings.Join(nodes, ","))
	}

	fmt.Printf("%s\n", string(y))
	return nil
}
