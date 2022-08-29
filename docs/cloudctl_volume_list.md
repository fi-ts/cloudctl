## cloudctl volume list

list all volumes

```
cloudctl volume list [flags]
```

### Options

```
  -h, --help               help for list
      --id string          volumeid to filter [optional]
      --only-unbound       show only unbound volumes that are not connected to any hosts, pv may be still present. [optional]
      --partition string   partition to filter [optional]
      --project string     project to filter [optional]
      --sort-by strings    sort by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: id|name|partition|project|tenant|usage
      --tenant string      tenant to filter [optional]
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

* [cloudctl volume](cloudctl_volume.md)	 - manage volume entities

