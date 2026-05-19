# cloudctl

Commandline client for "Kubernetes as a Service" and more!

<!-- TOC depthFrom:2 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [cloudctl](#cloudctl)
  - [Installation](#installation)
    - [Installation on Linux](#installation-on-linux)
    - [Installation on MacOS](#installation-on-macos)
    - [Installation on Windows](#installation-on-windows)
    - [cloudctl update](#cloudctl-update)
  - [Usage](#usage)
    - [Login](#login)
    - [Get currently logged in user](#get-currently-logged-in-user)
  - [HowTo](#howto)
    - [List Clusters](#list-clusters)
    - [Create Project](#create-project)
    - [Create Cluster](#create-cluster)
    - [Download Kubeconfig](#download-kubeconfig)
    - [Delete your cluster](#delete-your-cluster)
    - [Managing ip addresses](#managing-ip-addresses)
  - [Billing](#billing)
    - [Time Windows](#time-windows)
    - [Usage Calculation](#usage-calculation)
    - [Subcommands](#subcommands)
    - [CSV Export](#csv-export)
    - [Cost Calculation](#cost-calculation)
    - [Output Formats](#output-formats)
  - [S3](#s3)
    - [Configuring the minio mc client](#configuring-the-minio-mc-client)
    - [Configuring s3cmd](#configuring-s3cmd)
  - [Advanced Usage](#advanced-usage)
    - [Use token for existing Cluster](#use-token-for-existing-cluster)

<!-- /TOC -->

## Installation

Download locations:

- [cloudctl-linux-amd64](https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-linux-amd64)
- [cloudctl-darwin-amd64](https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-darwin-amd64)
- [cloudctl-darwin-arm64](https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-darwin-arm64)
- [cloudctl-windows-amd64](https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-windows-amd64)

[![Packaging status](https://repology.org/badge/vertical-allrepos/fits-cloudctl.svg)](https://repology.org/project/fits-cloudctl/versions)

### Installation on Linux

```bash
curl -LO https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-linux-amd64
chmod +x cloudctl-linux-amd64
sudo mv cloudctl-linux-amd64 /usr/local/bin/cloudctl
```

### Installation on MacOS

For x86 based Macs:

```bash
curl -LO https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-darwin-amd64
chmod +x cloudctl-darwin-amd64
sudo mv cloudctl-darwin-amd64 /usr/local/bin/cloudctl
```

For Apple Silicon (M1) based Macs:

```bash
curl -LO https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-darwin-arm64
chmod +x cloudctl-darwin-arm64
sudo mv cloudctl-darwin-arm64 /usr/local/bin/cloudctl
```

### Usage with Nix on Linux or MacOS

`fits-cloudctl` is packaged in [nixpkgs](https://github.com/NixOS/nixpkgs) and
can be installed using the [Nix Package Manager](https://nixos.org/) on Linux,
MacOS and NixOS.

```bash
$ nix-shell -p fits-cloudctl
```

The package can also be installed eg. with `nix-env -i fits-cloudctl`.

### Installation on Windows

```bash
curl -LO https://github.com/fi-ts/cloudctl/releases/latest/download/cloudctl-windows-amd64
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

Login, issue token for cloud and cluster access.

First you need to create a file in your home directory:

`~/.cloudctl/config.yaml`

```yaml
---
current: prod
contexts:
  prod:
    url: https://api.somedomain.example/cloud
    issuer_url: https://dex.somedomain.example
    client_id: my-client-id
    client_secret: my-secret
```

Optional you can specify `issuer_type: generic` if you use other issuers as Dex, e.g. Keycloak (this will request scopes `openid,profile,email`):
```yaml
contexts:
  prod:
    url: https://api.somedomain.example/cloud
    issuer_url: https://keycloak.somedomain.example
    issuer_type: generic
    client_id: my-client-id
    client_secret: my-secret
```

If you must specify special scopes for your issuer, you can use `custom_scopes`:
```yaml
contexts:
  prod:
    url: https://api.somedomain.example/cloud
    issuer_url: https://keycloak.somedomain.example
    custom_scopes: roles,openid,profile,email
    client_id: my-client-id
    client_secret: my-secret
```

Then you can login:

```bash
cloudctl login
```

A Browser window will open and you are prompted to select your backend.

- Choose the login for your organization and type your login credentials
- Push green button: "Grant Access"

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
  --partition <partition id> \
  --description "banking cluster for project banking next generation" \
  --minsize 2 \
  --maxsize 2

UID                                   NAME     VERSION  PARTITION  DOMAIN                               OPERATION  PROGRESS          APISERVER  CONTROL  NODES  SYSTEM  SIZE   AGE
1d8636d7-dadb-11e9-9e70-8ebea97dd3a9  banking  1.14.3   nbg-w8101  banking.pd25ml.cluster.somedomain.example  Succeeded  0% [Create]                                          2/2    1m

after ~7min:

cloudctl cluster ls
UID                                   NAME     VERSION  PARTITION  DOMAIN                               OPERATION  PROGRESS          APISERVER  CONTROL  NODES  SYSTEM  SIZE   AGE
1d8636d7-dadb-11e9-9e70-8ebea97dd3a9  banking  1.14.3   nbg-w8101  banking.pd25ml.cluster.somedomain.example  Succeeded  100% [Reconcile]  True       True     True   True    2/2  9m
```

Remember the cluster UID for further references.

You can list possible input options for the cluster create command via (some of them are defaulted, so you do not have to define all of them):

```bash
cloudctl cluster inputs
    firewallimages:
      - firewall-2.0.20200331
      - firewall-ubuntu-2.0.20200331
    firewalltypes:
      - c1-xlarge-x86
      - s2-xlarge-x86
    kubernetesversions:
      - 1.15.10
      - 1.15.11
      - 1.16.7
      - 1.16.8
      - 1.16.9
      - 1.17.3
      - 1.17.4
      - 1.17.5
    machineimages:
      - name: ubuntu
        version: "19.10"
    machinetypes:
      - c1-xlarge-x86
      - s2-xlarge-x86
    partitionconstraints:
        fel-wps101:
            networks:
              - internet
              - mpls-fits
              - ...
        nbg-w8101:
            networks:
              - internet
              - mpls-fits
              - ...
    partitions:
      - fel-wps101
      - nbg-w8101
```

### Download Kubeconfig

In order to be able to download the kubeconfig the cluster must have reached the APISERVER=True state.
This can be checked with subsequent `cloudctl cluster ls` calls, or even more convenient `watch cloudctl cluster ls`.

```bash
cloudctl cluster kubeconfig <cluster UID> > banking.kubeconfig

kubectl --kubeconfig ./banking.kubeconfig get nodes

```

### Delete your cluster

When you do not need your cluster anymore you can delete your cluster, to do so you get asked two questions to be sure you delete the correct cluster.
The first question asks you for the first part of the clusterID. If you clusterID looks like: `b5e24862-3cc2-4145-bfa4-ae4af102f965` the first part is up to the first `-` which is in this case `b5e24862`. The second question is the name of the cluster. If both was correct, your cluster will be deleted, if not the deletion is not triggered.

```bash
cloudctl cluster rm <cluster UID>
  UID                                   TENANT  PROJECT                               NAME        VERSION  PARTITION  OPERATION   PROGRESS      API   CONTROL  NODES  SYSTEM  SIZE  AGE
  9b86273a-0ab1-11ea-8057-9ad8c07d0e04  fits    b5e24862-3cc2-4145-bfa4-ae4af102f965  s3-cluster  1.14.3   nbg-w8101  Processing  63% [Delete]  True  False    True   False   1/1   17m 22s
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

`cloudctl billing` provides usage and cost tracking for all platform resources. Every billing query operates on a time window defined by `--from` and `--to`.

### Time Windows

| Flag | Default | Description |
|------|---------|-------------|
| `--from` | Start of current month | Beginning of the accounting window |
| `--to` | Now | End of the accounting window |
| `--time-format` | `2006-01-02` | Go time layout for parsing `--from` and `--to` |

```bash
# Current month (default)
cloudctl billing container -t mytenant

# Specific date range
cloudctl billing container -t mytenant --from 2024-01-01 --to 2024-01-31

# With custom time format including hours
cloudctl billing container -t mytenant \
  --from "2024-01-01 08:00:00" \
  --to "2024-01-31 17:00:00" \
  --time-format "2006-01-02 15:04:05"
```

### Usage Calculation

Usage is calculated as the integral of resource allocation over the time window. When resource limits change mid-window, the calculation tracks each step:

**Example**: A pod starts at 12:00 with 100m CPU, resource limits change to 200m at 12:30, time window is 12:00-13:00.

- CPU seconds = `1800s * 100m + 1800s * 200m = 540 CPU-seconds`
- Displayed as hours: `540s / 3600 = 0.15 CPU-hours`

### Subcommands

#### Discover Projects

List all projects that have billing data within a time period. Useful for narrowing down larger queries.

```bash
cloudctl billing projects --from 2024-01-01 --to 2024-01-31
```

#### Container Billing

Track CPU and memory usage of containers (pods).

```bash
cloudctl billing container -t mytenant -p <project-id>
cloudctl billing container -t mytenant -c <cluster-id> -n <namespace>
cloudctl billing container -t mytenant --annotations team=backend
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--namespace` | `-n` | Namespace to account |
| `--annotations` | | Filter by annotations |
| `--csv` | | Export as CSV |

Output fields include pod name, container name, image, lifetime, CPU (core * seconds), and memory (Gi * h).

#### Cluster Billing

Track cluster lifetime and worker group information.

```bash
cloudctl billing cluster -t mytenant
cloudctl billing cluster -t mytenant -p <project-id> -c <cluster-id>
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--csv` | | Export as CSV |

Output fields include partition, cluster name, start/end times, lifetime, and worker group details.

#### IP Billing

Track IP address allocation and lifetime.

```bash
cloudctl billing ip -t mytenant -p <project-id>
cloudctl billing ip -t mytenant --annotations env=production
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--annotations` | | Filter by annotations |
| `--csv` | | Export as CSV |

#### Network Traffic Billing

Track incoming, outgoing, and total network traffic in Gigabytes.

```bash
cloudctl billing network-traffic -t mytenant -p <project-id>
cloudctl billing network-traffic -t mytenant -c <cluster-id> --device eth0
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--device` | | Network device to account |
| `--csv` | | Export as CSV |

Output fields include device, incoming (Gi), outgoing (Gi), total (Gi), and lifetime.

#### S3 Billing

Track S3 object storage usage per bucket.

```bash
cloudctl billing s3 -t mytenant -p <project-id>
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--csv` | | Export as CSV |

Output fields include partition, user, bucket name/ID, object count, and storage (Gi * h).

#### Volume Billing

Track persistent volume (block storage) capacity usage.

```bash
cloudctl billing volume -t mytenant -p <project-id>
cloudctl billing volume -t mytenant -c <cluster-id> -n <namespace>
cloudctl billing volume -t mytenant --annotations team=backend
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--namespace` | `-n` | Namespace to account |
| `--annotations` | | Filter by annotations |
| `--csv` | | Export as CSV |

Output fields include volume UUID, name, storage type (e.g. fast, standard), capacity (Gi * h), and lifetime.

#### PostgreSQL Billing

Track CPU, memory, and storage usage of managed PostgreSQL instances.

```bash
cloudctl billing postgres -t mytenant -p <project-id>
cloudctl billing postgres -t mytenant --uuid <postgres-uuid>
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--uuid` | | PostgreSQL instance UUID |
| `--annotations` | | Filter by annotations |
| `--csv` | | Export as CSV |

Output fields include PostgreSQL ID, description, CPU (core * seconds), memory (Gi * h), storage (Gi * h), and lifetime.

#### Machine Billing

Track bare metal machine allocation and lifetime.

```bash
cloudctl billing machine -t mytenant -p <project-id>
cloudctl billing machine -t mytenant --machine-id <id> --size-id c1-xlarge-x86
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--machine-id` | | Specific machine ID |
| `--size-id` | | Machine size to filter |
| `--partition-id` | | Partition to filter |

#### Machine Reservation Billing

Track reserved machine slots and their utilization.

```bash
cloudctl billing machine-reservation -t mytenant -p <project-id>
cloudctl billing machine-reservation -t mytenant --size-id c1-xlarge-x86 --order tenant,project,size
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--id` | | Reservation ID |
| `--size-id` | | Machine size to filter |
| `--partition-id` | | Partition to filter |
| `--order` | | Sort by columns (e.g. `tenant,project,size,partition`) |

#### Product Option Billing

Track product option lifetime and usage.

```bash
cloudctl billing product-option -t mytenant -p <project-id>
cloudctl billing product-option -t mytenant -c <cluster-id> --id <option-id>
```

| Flag | Short | Description |
|------|-------|-------------|
| `--tenant` | `-t` | Tenant to account |
| `--project-id` | `-p` | Project to account |
| `--cluster-id` | `-c` | Cluster to account |
| `--id` | | Product option ID |

### CSV Export

Most billing commands support `--csv` to produce CSV output, which can be redirected to a file:

```bash
cloudctl billing container -t mytenant --csv > containers.csv
cloudctl billing volume -t mytenant -p <project-id> --csv > volumes.csv
```

CSV export is available for: container, cluster, ip, network-traffic, s3, volume, and postgres.

### Cost Calculation

When cost environment variables are configured, accumulated totals in the output include cost estimates in EUR. Set these via environment variables (prefixed with `CLOUDCTL_`) or in the cloudctl config:

| Variable | Applies To | Unit |
|----------|-----------|------|
| `costs-cpu-hour` | Container, PostgreSQL | EUR per CPU core hour |
| `costs-memory-gi-hour` | Container, PostgreSQL | EUR per GiB memory hour |
| `costs-storage-gi-hour` | S3, Volume, PostgreSQL | EUR per GiB storage hour |
| `costs-hour` | Cluster, IP, Machine, Machine Reservation, Product Option | EUR per hour |
| `costs-incoming-network-traffic-gi` | Network Traffic | EUR per GiB incoming |
| `costs-outgoing-network-traffic-gi` | Network Traffic | EUR per GiB outgoing |
| `costs-total-network-traffic-gi` | Network Traffic | EUR per GiB total |

### Output Formats

All billing commands support global output format options:

```bash
cloudctl billing container -t mytenant -o wide       # All columns
cloudctl billing container -t mytenant -o json        # JSON output
cloudctl billing container -t mytenant -o yaml        # YAML output
cloudctl billing container -t mytenant -o markdown    # Markdown table
cloudctl billing container -t mytenant --no-headers   # Omit column headers
```

## S3

You can manage S3 storage using `cloudctl` when S3 is configured in your metal stack control plane.

To list the available S3 partitions in your control plane, issue the following command:

```bash
$ cloudctl s3 partitions
NAME        ENDPOINT
fel-wps101  https://s3.test-01-fel-wps101.somedomain.example
```

In this case, the partition `fel-wps101` offers S3 storage. You can now create an S3 user to get storage access:

```bash
$ cloudctl s3 create --id my-user --project dc565451-3864-4355-bef5-080a9d0e4068 --partition fel-wps101 -n "My User"
accesskey: 3ZA4D7NFT1K6UB1N2ON1
name: My User
email: null
endpoint: https://s3.test-01-fel-wps101.somedomain.example
maxbuckets: 1000
id: my-user
partition: fel-wps101
project: dc565451-3864-4355-bef5-080a9d0e4068
secretkey: kEZ8fV1odMa9SzgrRlW9HtwB4yAqYITd4hM4NJTT
tenant: fits
```

After that, you can configure an S3 client to access the storage with this user.

If you need to look up the user at a later point in time again, you can use the describe command:

```bash
$ cloudctl s3 describe --name test --partition fel-wps101
accesskey: 3ZA4D7NFT1K6UB1N2ON1
name: My User
email: null
endpoint: https://s3.test-01-fel-wps101.somedomain.example
maxbuckets: 1000
id: my-user
partition: fel-wps101
project: dc565451-3864-4355-bef5-080a9d0e4068
secretkey: kEZ8fV1odMa9SzgrRlW9HtwB4yAqYITd4hM4NJTT
tenant: fits
```

Or if you want to delete the user again, run the delete command:

```bash
$ cloudctl s3 delete --name test --partition fel-wps101
endpoint: https://s3.test-01-fel-wps101.somedomain.example
id: my-user
partition: fel-wps101
project: dc565451-3864-4355-bef5-080a9d0e4068
tenant: fits
```

### Configuring the minio mc client

the command: `cloudctl s3 describe --for-client minio|s3cmd` will echo the required cli for the requested flavour.

```bash
$ cloudctl s3 describe --for-client minio --id test --partition=fel-wps101 --project=4fe217b4-3b3d-413e-87fc-fb89054cc70c
mc alias set test https://s3.test-01-fel-wps101.somedomain.example <your access key> <your secret key>

$ mc mb test/testbucket
Bucket created successfully `test/testbucket`.

$ mc cp ./README.md test/testbucket
./README.md:                       4.05 KiB / 4.05 KiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 100.00% 2.23 KiB/s 1s

$ mc ls test/testbucket
[2020-04-07 09:34:48 CEST]  4.0KiB README.md
```

### Configuring s3cmd

```bash
export AWS_ACCESS_KEY_ID= <your access key>
export AWS_SECRET_ACCESS_KEY=<your secret key>

$ cloudctl s3 describe --for-client s3cmd --id test --partition=fel-wps101 --project=4fe217b4-3b3d-413e-87fc-fb89054cc70c
cat << EOF > ${HOME}/.s3cfg
[default]
access_key = 45F3GU4DYSSN958I0HI8
host_base = https://s3.test-01-fel-wps101.somedomain.example
host_bucket = https://s3.test-01-fel-wps101.somedomain.example
secret_key = vee0Pa2ahgaec5Eitaucheedaij3oot9ahh2aeWe
EOF

s3cmd la
2020-04-07 07:34      4147   s3://test/README.md
```

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
