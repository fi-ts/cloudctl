## cloudctl tenant apply

applies one or more tenants from a given file

```
cloudctl tenant apply [flags]
```

### Options

```
      --bulk-output   prints results in a bulk at the end, the results are a list. default is printing results intermediately during apply, which causes single entities to be printed sequentially.
  -f, --file string   filename of the create or update request in yaml format, or - for stdin.
                      
                      Example:
                      $ cloudctl tenant describe tenant-1 -o yaml > tenant.yaml
                      $ vi tenant.yaml
                      $ # either via stdin
                      $ cat tenant.yaml | cloudctl tenant apply -f -
                      $ # or via file
                      $ cloudctl tenant apply -f tenant.yaml
                      	
      --force         skips security prompty for bulk operations
  -h, --help          help for apply
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

* [cloudctl tenant](cloudctl_tenant.md)	 - manage tenant entities

