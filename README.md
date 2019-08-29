# cloudctl

Commandline client for "Kubernetes as a Service" and more!

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
