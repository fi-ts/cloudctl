## cloudctl postgres create

create postgres

```
cloudctl postgres create [flags]
```

### Options

```
      --audit-logs             enable audit logs for the database (default true)
      --backup-config string   backup to use
      --buffer string          shared buffer for the database (default "64Mi")
      --cpu string             cpus for the database (default "500m")
      --description string     description of the database
  -h, --help                   help for create
      --labels strings         labels to add to that postgres database
      --maintenance strings    time specification of the automatic maintenance in the form Weekday:HH:MM-HH-MM [optional] (default [Sun:22:00-23:00])
      --partition string       partition where the database should be created
      --project string         project of the database
      --replicas int           replicas of the database (default 1)
      --sources strings        networks which should be allowed to connect in CIDR notation (default [0.0.0.0/0])
      --storage string         storage for the database (default "10Gi")
      --version string         version of the database (default "12")
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

