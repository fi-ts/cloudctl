# cloudctl

Commandline client for "Kubernetes as a Service" and more!

## Download from Blobstore

* [cloudctl-linux-amd64](https://blobstore.fi-ts.io/metal/cloudctl/cloudctl-linux-amd64)
* [cloudctl-darwin-amd64](https://blobstore.fi-ts.io/metal/cloudctl/cloudctl-darwin-amd64)
* [cloudctl-windows-amd64](https://blobstore.fi-ts.io/metal/cloudctl/cloudctl-windows-amd64)

## Usage

### Login

Login, issue token for cloud and cluster access
~~~~
cloudctl login
~~~~

Token will be written to default kubectl-config, e.g. ~/.kube/config

### Get currently logged in user

~~~~
cloudctl whoami
~~~~

Prints the username, that is currently logged in. This does not mean, that the token is still vaild.

### Use token for existing Cluster

1. You have downloaded your kubeconfig to the default location or inserted the cluster-config into your existing kubeconfig under the name "clustername".
2. issue token "cloudctl login", will be stored in config, get name with "cloudctl whoami"
3. assign your user credentials with your cluster in the context "contextname", see following paragraph 

If you want to use your token for "username" for your cluster "clustername" in the context "contextname" (existing or new) then you have to issue the following command:
~~~~
kubectl config set-context contextname --user username --cluster clustername [--namespace=mynamespace]
~~~~

This prepares your context "contextname" in a way, that your user credentials of user "username" are used with the cluster "clustername".
You can assign your user to multiple clusters.

This process has to be done only once. The next time you execute "cloudctl login", the token can be used for all contexts the user has been assigned to.


## HowTo

Current Scenario
* metal-Controlplane & Garden at GKE
* Seed-Partition fra-equ01
* cloudclt connects directly to metal-api and garden-controlplane

Currently you have to specify
~~~~
export CLOUDCTL_URL=https://api.metal-pod.dev/metal
export CLOUDCTL_HMAC=ytjd...
~~~~

You must always use the kubeconfig, pointing you to the garden-controlplane.
* Hint: you can fetch it with metal-deployment/control-plane/Makefile#fetch-kubeconfig-from-gcp
* Hint2: be sure to first 'docker pull registry.fi-ts.io/metal/metal-deployment-base:latest' 

List Clusters
~~~~
bin/cloudctl --kubeconfig gke-kube-config.yaml cluster list
~~~~

Create Project
~~~~
bin/cloudctl project create --kubeconfig gke-kube-config.yaml --name=banking
~~~~

Create Cluster
~~~~
bin/cloudctl cluster create --kubeconfig gke-kube-config.yaml --name=banking --owner=hans --partition=fra-equ01 --project=f4ea5de9-18fa-4df9-ad4c-61e21c57d03e --description="banking cluster for project banking next generation"
~~~~