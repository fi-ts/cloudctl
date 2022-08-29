## cloudctl billing cluster

look at cluster bills

### Synopsis


You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.


```
cloudctl billing cluster [flags]
```

### Options

```
  -c, --cluster-id string    the cluster to account
      --csv                  let the server generate a csv file
      --from string          the start time in the accounting window to look at (optional, defaults to start of the month
  -h, --help                 help for cluster
  -p, --project-id string    the project to account
  -t, --tenant string        the tenant to account
      --time-format string   the time format used to parse the arguments 'from' and 'to' (default "2006-01-02")
      --to string            the end time in the accounting window to look at (optional, defaults to current system time)
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

* [cloudctl billing](cloudctl_billing.md)	 - lookup resource consumption of your cloud resources

