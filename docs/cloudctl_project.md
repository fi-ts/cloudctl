## cloudctl project

manage project entities

### Synopsis

a project organizes cloud resources regarding tenancy, quotas, billing and authentication

### Options

```
  -h, --help   help for project
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
* [cloudctl project apply](cloudctl_project_apply.md)	 - applies one or more projects from a given file
* [cloudctl project create](cloudctl_project_create.md)	 - creates the project
* [cloudctl project delete](cloudctl_project_delete.md)	 - deletes the project
* [cloudctl project describe](cloudctl_project_describe.md)	 - describes the project
* [cloudctl project edit](cloudctl_project_edit.md)	 - edit the project through an editor and update
* [cloudctl project list](cloudctl_project_list.md)	 - list all projects
* [cloudctl project update](cloudctl_project_update.md)	 - updates the project

