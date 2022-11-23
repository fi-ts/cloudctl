## cloudctl cluster create

create a cluster

```
cloudctl cluster create [flags]
```

### Options

```
      --allowprivileged                allow privileged containers the cluster.
      --audit string                   audit logging of cluster API access; can be off, on (default) or splunk (logging to a predefined or custom splunk endpoint). [optional] (default "on")
      --autoupdate-kubernetes          enables automatic updates of the kubernetes patch version of the cluster [optional]
      --autoupdate-machineimages       enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes! [optional]
      --cni string                     the network plugin used in this cluster, defaults to calico. please note that cilium support is still Alpha and we are happy to receive feedback. [optional]
      --cri string                     container runtime to use, only docker|containerd supported as alternative actually. [optional]
      --default-storage-class string   set default storage class to given name, must be one of the managed storage classes
      --description string             description of the cluster. [optional]
      --draintimeout duration          period (e.g. "3h") after which a draining node will be forcefully deleted. [optional]
      --egress strings                 static egress ips per network, must be in the form <network>:<ip>; e.g.: --egress internet:1.2.3.4,extnet:123.1.1.1 --egress internet:1.2.3.5 [optional]
      --encrypted-storage-classes      enables the deployment of encrypted duros storage classes into the cluster. please refer to the user manual to properly use volume encryption. [optional]
      --external-networks strings      external networks of the cluster
      --firewallcontroller string      version of the firewall-controller to use. [optional]
      --firewallimage string           machine image to use for the firewall. [optional]
      --firewalltype string            machine type to use for the firewall. [optional]
      --healthtimeout duration         period (e.g. "24h") after which an unhealthy node is declared failed and will be replaced. [optional]
  -h, --help                           help for create
      --labels strings                 labels of the cluster
      --logacceptedconns               also log accepted connections on the cluster firewall [optional]
      --machineimage string            machine image to use for the nodes, must be in the form of <name>-<version> [optional]
      --machinetype string             machine type to use for the nodes. [optional]
      --max-pods-per-node string       set number of maximum pods per node (default: 510). Lower numbers allow for more node per cluster. [optional]
      --maxsize int32                  maximal workers of the cluster. (default 1)
      --maxsurge string                max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster. (default "1")
      --maxunavailable string          max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster. (default "0")
      --minsize int32                  minimal workers of the cluster. (default 1)
      --name string                    name of the cluster, max 10 characters. [required]
      --partition string               partition of the cluster. [required]
      --project string                 project where this cluster should belong to. [required]
      --purpose string                 purpose of the cluster, can be one of production|development|evaluation|infrastructure. SLA is only given on production clusters. [optional] (default "evaluation")
      --reversed-vpn                   enables usage of reversed-vpn instead of konnectivity tunnel for worker connectivity. [optional]
      --seed string                    name of seed where this cluster should be scheduled. [optional]
      --version string                 kubernetes version of the cluster. defaults to latest available, check cluster inputs for possible values. [optional]
```

### Options inherited from parent commands

```
      --api-token string       api token to authenticate. Can be specified with CLOUDCTL_API_TOKEN environment variable.
      --api-url string         api server address. Can be specified with CLOUDCTL_API_URL environment variable.
      --debug                  debug output
      --force-color            force colored output even without tty
      --kubeconfig string      Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.
      --no-headers             omit headers in tables
      --order string           order by (comma separated) column(s)
  -o, --output-format string   output format (table|wide|markdown|json|yaml|template), wide is a table with more columns. (default "table")
      --template string        output template for template output-format, go template format.
                               	For property names inspect the output of -o json for reference.
                               	Example for clusters:
                               
                               	cloudctl cluster ls -o template --template "{{ .ID }} {{ .Name }}"
                               
                               	
      --yes-i-really-mean-it   skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)
```

### SEE ALSO

* [cloudctl cluster](cloudctl_cluster.md)	 - manage clusters

