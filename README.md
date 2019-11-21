# cloudctl

Commandline client for "Kubernetes as a Service" and more!

## Installation

Download locations:

* [cloudctl-linux-amd64](https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-linux-amd64)
* [cloudctl-darwin-amd64](https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-darwin-amd64)
* [cloudctl-windows-amd64](https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-windows-amd64)

### Installation on Linux

```bash
curl -LO https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-linux-amd64
chmod +x cloudctl-linux-amd64
sudo mv cloudctl-linux-amd64 /usr/local/bin/cloudctl
```

### Installation on MacOS

```bash
curl -LO https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-darwin-amd64
chmod +x cloudctl-darwin-amd64
sudo mv cloudctl-darwin-amd64 /usr/local/bin/cloudctl
```

### Installation on Windows

```bash
curl -LO https://blobstore.fi-ts.io/cloud-native/cloudctl/cloudctl-windows-amd64
copy cloudctl-windows-amd64 cloudctl.exe
```

### cloudctl update

In order to keep your local `cloudctl` installation up to date, you can update the binary like this:

```bash
cloudctl update check
latest version:2019-09-20T08:48:07Z
local  version:2019-09-21T18:52:07Z
cloudctl is not up to date

cloudctl update do
# a download with progress bar starts and replaces the binary. If the binary has root permissions please execute
sudo cloudctl update do
# instead
```

## Usage

### Login

Login, issue token for cloud and cluster access

```bash
cloudctl login
```

A Browser window will open and you are prompted to select your backend.

* Choose "Log in with OpenLDAP Demo (TEST)"
* Push green button: "Grant Access"

Token will be written to default kubectl-config, e.g. ~/.kube/config

Then you can close the browser window.

### Get currently logged in user

```bash
cloudctl whoami
```

Prints the username, that is currently logged in. This does not mean, that the token is still valid.

## HowTo

### List Clusters

```bash
cloudctl cluster ls
```

### Create Project

```bash
cloudctl project create --name banking --description "Banking Cluster"

cloudctl project ls
UID                                   NAME     DESCRIPTION
25195ae3-8e02-4b56-ba36-d4b1f94bc17e  banking  Banking Cluster
```

Remember project UID for cluster creation.

### Create Cluster

```bash
cloudctl cluster create \
  --name banking \
  --project <project UID> \
  --description "banking cluster for project banking next generation" \
  --minsize 2 \
  --maxsize 2

UID                                   NAME     VERSION  PARTITION  DOMAIN                               OPERATION  PROGRESS          APISERVER  CONTROL  NODES  SYSTEM  SIZE   AGE
1d8636d7-dadb-11e9-9e70-8ebea97dd3a9  banking  1.14.3   nbg-w8101  banking.pd25ml.cluster.metal-pod.io  Succeeded  0% [Create]                                          2/2    1m

after ~7min:

cloudctl cluster ls
UID                                   NAME     VERSION  PARTITION  DOMAIN                               OPERATION  PROGRESS          APISERVER  CONTROL  NODES  SYSTEM  SIZE   AGE
1d8636d7-dadb-11e9-9e70-8ebea97dd3a9  banking  1.14.3   nbg-w8101  banking.pd25ml.cluster.metal-pod.io  Succeeded  100% [Reconcile]  True       True     True   True    2/2  9m
```

Remember the cluster UID for further references.

### Download Kubeconfig

In order to be able to download the kubeconfig the cluster must have reached the APISERVER=True state.
This can be checked with subsequent `cloudctl cluster ls` calls, or even more convenient `watch cloudctl cluster ls`.

```bash
cloudctl cluster credentials <cluster UID> > banking.kubeconfig

kubectl --kubeconfig ./banking.kubeconfig get nodes

```

### Delete your cluster

When you do not need your cluster anymore you can delete your cluster, to do so you get asked two questions to be sure you delete the correct cluster.
The first question asks you for the first part of the clusterID. If you clusterID looks like: `b5e24862-3cc2-4145-bfa4-ae4af102f965` the first part is up to the first `-` which is in this case `b5e24862`. The second question is the name of the cluster. If both was correct, your cluster will be deleted, if not the deletion is not triggered.

```bash
cloudctl cluster rm <cluster UID>
  UID                                   TENANT  PROJECT                               NAME        VERSION  PARTITION  OPERATION   PROGRESS      API   CONTROL  NODES  SYSTEM  SIZE  AGE
  9b86273a-0ab1-11ea-8057-9ad8c07d0e04  fits    b5e24862-3cc2-4145-bfa4-ae4af102f965  s3-cluster  1.14.3   fra-equ01  Processing  63% [Delete]  True  False    True   False   1/1   17m 22s
Please answer some security questions to delete this cluster
first part of clusterID:9b86273a
Clustername:s3-cluster
```

### Managing ip addresses

Ingress ip addresses in Kubernetes are generated automatically from an ip address pool
when a service of LoadBalancer type is created. The default address pool is the internet address pool,
so applications can be reached from the internet.

In order to make applications accessible from the "internal" MPLS network, the appropriate address pool
must be specified. The available pools can be found with:

```bash
cloudctl cluster describe <cluster UID>

(...)
spec:
(...)
    cloud:
(...)
        metal:
(...)
            networks:
                additional:
                  - internet-nbg-w8101
                  - mpls-nbg-w8101-fits-dev

```

The pool is added as annotation to the service definition:

```bash
apiVersion: v1
kind: Service
metadata:
  name: <service name>
  labels:
    name: <service name>
    app: <application name>
  annotations:
    metallb.universe.tf/address-pool: mpls-nbg-w8101-fits-dev-ephemeral
spec:
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
  selector:
     name: <pod name>
     app: <application name>
```

(Note the attached -ephemeral at the end of the pool name.)

After applying the service definition, its ip address can be found with kubectl get services:

```bash
kubectl --kubeconfig banking.kubeconfig get services
NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
db                    ClusterIP      10.244.28.130   <none>          5432/TCP       16s
kubernetes            ClusterIP      10.244.16.1     <none>          443/TCP        24m
redis                 ClusterIP      10.244.21.227   <none>          6379/TCP       16s
result-service        LoadBalancer   10.244.19.185   212.34.89.84    80:31643/TCP   16s
voting-service        LoadBalancer   10.244.27.21    212.34.89.85    80:31866/TCP   16s
voting-service-mpls   LoadBalancer   10.244.19.97    100.127.129.3   80:32708/TCP   16s
```

To list all ip addresses assigned to the current project, use

```bash
cloudctl ip list --project <project UID>
```

The output could look like this:

```bash
cloudctl ip list --project 9725892b-a830-4ed9-b16a-75e2409c8316
  IP             TYPE       NAME                            NETWORK                               PROJECT                               TAGS
  10.2.0.3       ephemeral  shoot--pqpgh...-firewall-ebba7  underlay-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  10.3.60.1      ephemeral  shoot--pqpgh...-firewall-ebba7  4ff30487-9496-4770-8b88-38406ecd9bba  9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  10.3.60.2      ephemeral  shoot--pqpgh...8774f56d4-glhsw  4ff30487-9496-4770-8b88-38406ecd9bba  9725892b-a830-4ed9-b16a-75e2409c8316  machine:8c4eaa00-7187-11e9-8000-efbeaddeefbe
  100.127.129.2  ephemeral  shoot--pqpgh...-firewall-ebba7  mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  100.127.129.3  ephemeral  metallb-4e8d5                   mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/voting-service-mpls
  212.34.89.53   ephemeral  shoot--pqpgh...-firewall-ebba7  internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  212.34.89.54   ephemeral  metallb-a756a                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/kube-system/vpn-shoot
  212.34.89.55   ephemeral  metallb-a7a35                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/kube-system/addons-nginx-ingress-controller
  212.34.89.80   ephemeral  metallb-4e72b                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/result-service
  212.34.89.81   ephemeral  metallb-4eb0d                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/voting-service
```

To make an ip address static, use the command cloudctl ip static:

```bash
cloudctl ip static <ip address>
```

Static ip addresses are shown differently in the output of cloudctl ip list:

```bash
  IP             TYPE    NAME           NETWORK                  PROJECT                               TAGS
(...)
  100.127.129.3  static     metallb-4e8d5                   mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/voting-service-mpls
```

To bind a static IP to a service, the address gets added explicitly to the specs of the LoadBalancer service:

```bash
apiVersion: v1
kind: Service
metadata:
  name: <service name>
  labels:
    name: <service name>
    app: <application name>
spec:
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
  selector:
     name: <pod name>
     app: <application name>
  loadBalancerIP: 100.127.129.3
```

The same static ip address can be bound to services in different clusters by specifying it in the appropriate
service definitions. The output of cloudctl ip list then shows both services for the same ip address:

```bash
cloudctl ip list --project 9725892b-a830-4ed9-b16a-75e2409c8316
  IP             TYPE       NAME                            NETWORK                               PROJECT                               TAGS
  10.2.0.3       ephemeral  shoot--pqpgh...-firewall-ebba7  underlay-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  10.2.0.5       ephemeral  shoot--pqpgh...-firewall-2af68  underlay-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:00000000-beef-beef-0006-efbeaddeefbe
  10.3.60.1      ephemeral  shoot--pqpgh...-firewall-ebba7  4ff30487-9496-4770-8b88-38406ecd9bba  9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  10.3.60.2      ephemeral  shoot--pqpgh...8774f56d4-glhsw  4ff30487-9496-4770-8b88-38406ecd9bba  9725892b-a830-4ed9-b16a-75e2409c8316  machine:8c4eaa00-7187-11e9-8000-efbeaddeefbe
  10.3.96.1      ephemeral  shoot--pqpgh...-firewall-2af68  0f573b0b-bbe4-4d80-a767-3b522ab0fb08  9725892b-a830-4ed9-b16a-75e2409c8316  machine:00000000-beef-beef-0006-efbeaddeefbe
  10.3.96.2      ephemeral  shoot--pqpgh...84cd5df9d-q45lz  0f573b0b-bbe4-4d80-a767-3b522ab0fb08  9725892b-a830-4ed9-b16a-75e2409c8316  machine:00000000-beef-beef-0012-efbeaddeefbe
  100.127.129.2  ephemeral  shoot--pqpgh...-firewall-ebba7  mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  100.127.129.3  static     metallb-4e8d5                   mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  service:1c1e0e4c-024c-11ea-8b7d-e6272ba300ae/default/voting-service-mpls
                                                                                                                                        service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/voting-service-mpls
  100.127.129.4  ephemeral  shoot--pqpgh...-firewall-2af68  mpls-nbg-w8101-fits-dev               9725892b-a830-4ed9-b16a-75e2409c8316  machine:00000000-beef-beef-0006-efbeaddeefbe
  212.34.89.53   ephemeral  shoot--pqpgh...-firewall-ebba7  internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:71d4ec00-7107-11e9-8000-efbeaddeefbe
  212.34.89.54   ephemeral  metallb-a756a                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/kube-system/vpn-shoot
  212.34.89.55   ephemeral  metallb-a7a35                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/kube-system/addons-nginx-ingress-controller
  212.34.89.78   ephemeral  shoot--pqpgh...-firewall-2af68  internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  machine:00000000-beef-beef-0006-efbeaddeefbe
  212.34.89.79   ephemeral  metallb-fc775                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:1c1e0e4c-024c-11ea-8b7d-e6272ba300ae/kube-system/addons-nginx-ingress-controller
  212.34.89.80   ephemeral  metallb-4e72b                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/result-service
  212.34.89.81   ephemeral  metallb-4eb0d                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:deb74391-0245-11ea-8b7d-e6272ba300ae/default/voting-service
  212.34.89.82   ephemeral  metallb-fd499                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:1c1e0e4c-024c-11ea-8b7d-e6272ba300ae/kube-system/vpn-shoot
  212.34.89.84   ephemeral  metallb-ace53                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:1c1e0e4c-024c-11ea-8b7d-e6272ba300ae/default/result-service
  212.34.89.85   ephemeral  metallb-ad01d                   internet-nbg-w8101                    9725892b-a830-4ed9-b16a-75e2409c8316  service:1c1e0e4c-024c-11ea-8b7d-e6272ba300ae/default/voting-service
```

Static ip addresses are assigned to the project and survive deletion of individual clusters. To free a static ip address use the command cloudctl ip delete:

```bash
cloudctl ip delete 100.127.129.3
  IP             TYPE    NAME           NETWORK                  PROJECT                               TAGS
  100.127.129.3  static  metallb-4e8d5  mpls-nbg-w8101-fits-dev  9725892b-a830-4ed9-b16a-75e2409c8316
```

Static ip addresses must freed before their project can be deleted.

## Billing

The usage is calculated always withing a time window. The beginning of the time window can be specified by `--from` and if required `--to` specifies the end of the time window to look at. The end defaults to `now`.

Example calculation:

given

* pod starts at 12am with 100m cpu resource limits set
* time window between 12am and 1pm
* pod resource limits get modified from 100m to 200m at 12:30am

results

* pod lifetime is 1hour
* cpu seconds in the time window is the integral of a step function:
  `1800s * 100ms + 1800s * 200ms = 540000ms*s = 540s*s (=> 30min with cpu:100m and 30min with cpu:200m)`
* for the sake of readability, the output of cloudctl is made in hours: `540s*s/3600s => 0,15s*h`

### cluster

### container

### volumes



## Advanced Usage

### Use token for existing Cluster

1. You have downloaded your kubeconfig to the default location or inserted the cluster-config into your existing kubeconfig under the name "clustername".
2. issue token "cloudctl login", will be stored in config, get name with "cloudctl whoami"
3. assign your user credentials with your cluster in the context "contextname", see following paragraph

If you want to use your token for "username" for your cluster "clustername" in the context "contextname" (existing or new) then you have to issue the following command:

```bash
kubectl config set-context contextname --user username --cluster clustername [--namespace=mynamespace]
```

This prepares your context "contextname" in a way, that your user credentials of user "username" are used with the cluster "clustername".
You can assign your user to multiple clusters.

This process has to be done only once. The next time you execute "cloudctl login", the token can be used for all contexts the user has been assigned to.
