## cloudctl ip allocate

allocate a static IP address for your project that can be used for your cluster's service type load balancer

```
cloudctl ip allocate <ip> [flags]
```

### Options

```
      --description string   set description of the ip address [required]
  -h, --help                 help for allocate
      --name string          set name of the ip address [required]
      --network string       the network of the ip address [required]
      --project string       the project of the ip address [required]
      --specific-ip string   try allocating a specific ip address from a network [optional]
      --tags strings         set tags of the ip address [optional]
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

* [cloudctl ip](cloudctl_ip.md)	 - manage ips

