## cloudctl billing

lookup resource consumption of your cloud resources

### Options

```
  -h, --help   help for billing
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
* [cloudctl billing cluster](cloudctl_billing_cluster.md)	 - look at cluster bills
* [cloudctl billing container](cloudctl_billing_container.md)	 - look at container bills
* [cloudctl billing ip](cloudctl_billing_ip.md)	 - look at ip bills
* [cloudctl billing network-traffic](cloudctl_billing_network-traffic.md)	 - look at network traffic bills
* [cloudctl billing postgres](cloudctl_billing_postgres.md)	 - look at postgres bills
* [cloudctl billing projects](cloudctl_billing_projects.md)	 - discover projects within a given time period
* [cloudctl billing s3](cloudctl_billing_s3.md)	 - look at s3 bills
* [cloudctl billing volume](cloudctl_billing_volume.md)	 - look at volume bills

