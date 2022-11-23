## cloudctl ip

manage ips

### Synopsis

TODO

### Options

```
  -h, --help   help for ip
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
* [cloudctl ip allocate](cloudctl_ip_allocate.md)	 - allocate a static IP address for your project that can be used for your cluster's service type load balancer
* [cloudctl ip delete](cloudctl_ip_delete.md)	 - delete an ip
* [cloudctl ip list](cloudctl_ip_list.md)	 - list ips
* [cloudctl ip static](cloudctl_ip_static.md)	 - make an ephemeral ip static such that it won't be deleted if not used anymore

