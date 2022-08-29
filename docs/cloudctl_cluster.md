## cloudctl cluster

manage clusters

### Synopsis

TODO

### Options

```
  -h, --help   help for cluster
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

* [cloudctl](cloudctl.md)	 - a cli to manage cloud entities.
* [cloudctl cluster create](cloudctl_cluster_create.md)	 - create a cluster
* [cloudctl cluster delete](cloudctl_cluster_delete.md)	 - delete a cluster
* [cloudctl cluster describe](cloudctl_cluster_describe.md)	 - describe a cluster
* [cloudctl cluster dns-manifest](cloudctl_cluster_dns-manifest.md)	 - create a manifest for an ingress or service type loadbalancer, creating a DNS entry and valid certificate within your cluster domain
* [cloudctl cluster inputs](cloudctl_cluster_inputs.md)	 - get possible cluster inputs like k8s versions, etc.
* [cloudctl cluster issues](cloudctl_cluster_issues.md)	 - lists cluster issues, shows required actions explicitly when id argument is given
* [cloudctl cluster kubeconfig](cloudctl_cluster_kubeconfig.md)	 - get cluster kubeconfig
* [cloudctl cluster list](cloudctl_cluster_list.md)	 - list clusters
* [cloudctl cluster logs](cloudctl_cluster_logs.md)	 - get logs for the cluster
* [cloudctl cluster machine](cloudctl_cluster_machine.md)	 - list and access machines in the cluster
* [cloudctl cluster monitoring-secret](cloudctl_cluster_monitoring-secret.md)	 - returns the endpoint and access credentials to the monitoring of the cluster
* [cloudctl cluster reconcile](cloudctl_cluster_reconcile.md)	 - trigger cluster reconciliation
* [cloudctl cluster splunk-config-manifest](cloudctl_cluster_splunk-config-manifest.md)	 - create a manifest for a custom splunk configuration, overriding the default settings for splunk auditing
* [cloudctl cluster update](cloudctl_cluster_update.md)	 - update a cluster

