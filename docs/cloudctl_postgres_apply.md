## cloudctl postgres apply

apply postgres

```
cloudctl postgres apply [flags]
```

### Options

```
  -f, --file string   filename of the create or update request in yaml format, or - for stdin.
                      	Example postgres update:
                      
                      	# cloudctl postgres describe postgres1 -o yaml > postgres1.yaml
                      	# vi postgres1.yaml
                      	## either via stdin
                      	# cat postgres1.yaml | cloudctl postgres apply -f -
                      	## or via file
                      	# cloudctl postgres apply -f postgres1.yaml
                      	
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

* [cloudctl postgres](cloudctl_postgres.md)	 - manage postgres

