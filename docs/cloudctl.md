## cloudctl

a cli to manage cloud entities.

### Synopsis

with cloudctl you can manage kubernetes cluster, view networks et.al.

### Options

```
      --api-token string       api token to authenticate. Can be specified with CLOUDCTL_API_TOKEN environment variable.
      --api-url string         api server address. Can be specified with CLOUDCTL_API_URL environment variable.
      --debug                  debug output
      --force-color            force colored output even without tty
  -h, --help                   help for cloudctl
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

* [cloudctl billing](cloudctl_billing.md)	 - lookup resource consumption of your cloud resources
* [cloudctl cluster](cloudctl_cluster.md)	 - manage clusters
* [cloudctl completion](cloudctl_completion.md)	 - Generate the autocompletion script for the specified shell
* [cloudctl context](cloudctl_context.md)	 - manage cloudctl context
* [cloudctl dashboard](cloudctl_dashboard.md)	 - shows a live dashboard optimized for operation
* [cloudctl health](cloudctl_health.md)	 - show health information
* [cloudctl ip](cloudctl_ip.md)	 - manage ips
* [cloudctl login](cloudctl_login.md)	 - login user and receive token
* [cloudctl logout](cloudctl_logout.md)	 - logout user from OIDC SSO session
* [cloudctl markdown](cloudctl_markdown.md)	 - create markdown documentation
* [cloudctl postgres](cloudctl_postgres.md)	 - manage postgres
* [cloudctl project](cloudctl_project.md)	 - manage project entities
* [cloudctl s3](cloudctl_s3.md)	 - manage s3 entities
* [cloudctl tenant](cloudctl_tenant.md)	 - manage tenant entities
* [cloudctl update](cloudctl_update.md)	 - update the program
* [cloudctl version](cloudctl_version.md)	 - print the client and server version information
* [cloudctl volume](cloudctl_volume.md)	 - manage volume entities
* [cloudctl whoami](cloudctl_whoami.md)	 - shows current user

