## cloudctl cluster update

update a cluster

```
cloudctl cluster update <clusterid> [flags]
```

### Options

```
      --addlabels strings                      labels to add to the cluster
      --allowprivileged                        allow privileged containers the cluster, please add --yes-i-really-mean-it
      --audit string                           audit logging of cluster API access; can be off, on or splunk (logging to a predefined or custom splunk endpoint). (default "on")
      --autoupdate-kubernetes                  enables automatic updates of the kubernetes patch version of the cluster
      --autoupdate-machineimages               enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes!
      --default-storage-class string           set default storage class to given name, must be one of the managed storage classes
      --disable-custom-default-storage-class   if set to true, no default class is deployed, you have to set one of your storageclasses manually to default
      --draintimeout duration                  period (e.g. "3h") after which a draining node will be forcefully deleted. (0 = provider-default)
      --egress strings                         static egress ips per network, must be in the form <networkid>:<semicolon-separated ips>; e.g.: --egress internet:1.2.3.4;1.2.3.5 --egress extnet:123.1.1.1 [optional]. Use --egress none to remove all egress rules.
      --encrypted-storage-classes              enables the deployment of encrypted duros storage classes into the cluster. please refer to the user manual to properly use volume encryption.
      --external-networks strings              external networks of the cluster
      --firewallcontroller string              version of the firewall-controller to use.
      --firewallimage string                   machine image to use for the firewall.
      --firewalltype string                    machine type to use for the firewall.
      --healthtimeout duration                 period (e.g. "24h") after which an unhealthy node is declared failed and will be replaced. (0 = provider-default)
  -h, --help                                   help for update
      --logacceptedconns                       enables logging of accepted connections on the cluster firewall
      --machineimage string                    machine image to use for the nodes, must be in the form of <name>-<version> 
      --machinetype string                     machine type to use for the nodes.
      --maxsize int32                          maximal workers of the cluster.
      --maxsurge string                        max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.
      --maxunavailable string                  max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.
      --minsize int32                          minimal workers of the cluster.
      --purpose string                         purpose of the cluster, can be one of production|development|evaluation|infrastructure. SLA is only given on production clusters.
      --remove-workergroup                     if set, removes the targeted worker group
      --removelabels strings                   labels to remove from the cluster
      --reversed-vpn                           enables usage of reversed-vpn instead of konnectivity tunnel for worker connectivity.
      --seed string                            name of seed where this cluster should be scheduled.
      --version string                         kubernetes version of the cluster.
      --workerannotations strings              annotations of the worker group (syncs to kubernetes node resource after some time, too)
      --workergroup string                     the name of the worker group to apply updates to, only required when there are multiple worker groups.
      --workerlabels strings                   labels of the worker group (syncs to kubernetes node resource after some time, too)
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

