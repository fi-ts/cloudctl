## cloudctl project create

creates the project

```
cloudctl project create [flags]
```

### Options

```
      --annotation strings    add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar
      --bulk-output           when creating from file: prints results in a bulk at the end, the results are a list. default is printing results intermediately during creation, which causes single entities to be printed sequentially.
      --cluster-quota int32   cluster quota
      --description string    description of the project.
  -f, --file string           filename of the create or update request in yaml format, or - for stdin.
                              
                              Example:
                              $ cloudctl project describe project-1 -o yaml > project.yaml
                              $ vi project.yaml
                              $ # either via stdin
                              $ cat project.yaml | cloudctl project create -f -
                              $ # or via file
                              $ cloudctl project create -f project.yaml
                              	
  -h, --help                  help for create
      --ip-quota int32        ip quota
      --label strings         add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar
      --machine-quota int32   machine quota
      --name string           name of the project, max 10 characters.
      --tenant string         create project for given tenant
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

* [cloudctl project](cloudctl_project.md)	 - manage project entities

