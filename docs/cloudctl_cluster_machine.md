## cloudctl cluster machine

list and access machines in the cluster

### Options

```
  -h, --help   help for machine
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
* [cloudctl cluster machine console](cloudctl_cluster_machine_console.md)	 - console access a machine/firewall of the cluster
* [cloudctl cluster machine cycle](cloudctl_cluster_machine_cycle.md)	 - soft power cycle of a machine/firewall of the cluster
* [cloudctl cluster machine ls](cloudctl_cluster_machine_ls.md)	 - list machines of the cluster
* [cloudctl cluster machine reinstall](cloudctl_cluster_machine_reinstall.md)	 - reinstall OS image onto a machine/firewall of the cluster
* [cloudctl cluster machine reset](cloudctl_cluster_machine_reset.md)	 - hard power reset of a machine/firewall of the cluster
* [cloudctl cluster machine ssh](cloudctl_cluster_machine_ssh.md)	 - ssh access a machine/firewall of the cluster

