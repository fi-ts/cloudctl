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

A Browser window will open and you are prompted to select your authentification backend.
Choose "Log in with OpenLDAP Demo (TEST)"
Push green button: "Grant Access"

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

### Use your cluster

Now you are ready to use your Cluster with kubectl.

### Delete your cluster

When you do not need your cluster anymore you can delete your cluster:

```bash
cloudctl cluster rm <cluster UID>
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
