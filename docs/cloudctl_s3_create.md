## cloudctl s3 create

creates the s3

```
cloudctl s3 create [flags]
```

### Options

```
      --access-key string   specify the access key, otherwise will be generated
      --bulk-output         when creating from file: prints results in a bulk at the end, the results are a list. default is printing results intermediately during creation, which causes single entities to be printed sequentially.
  -f, --file string         filename of the create or update request in yaml format, or - for stdin.
                            
                            Example:
                            $ cloudctl s3 describe s3-1 -o yaml > s3.yaml
                            $ vi s3.yaml
                            $ # either via stdin
                            $ cat s3.yaml | cloudctl s3 create -f -
                            $ # or via file
                            $ cloudctl s3 create -f s3.yaml
                            	
  -h, --help                help for create
  -i, --id string           id of the s3 user [required]
      --max-buckets int     maximum number of buckets for the s3 user
  -n, --name string         name of s3 user, only for display
  -p, --partition string    name of s3 partition to create the s3 user in [required]
      --project string      id of the project that the s3 user belongs to [required]
      --secret-key string   specify the secret key, otherwise will be generated
  -t, --tenant string       create s3 for given tenant, defaults to logged in tenant
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

* [cloudctl s3](cloudctl_s3.md)	 - manage s3 entities

