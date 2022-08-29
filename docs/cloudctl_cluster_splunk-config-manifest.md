## cloudctl cluster splunk-config-manifest

create a manifest for a custom splunk configuration, overriding the default settings for splunk auditing

```
cloudctl cluster splunk-config-manifest [flags]
```

### Options

```
      --cabase64 string   the base64-encoded ca certificate (chain) for the splunk HEC endpoint
      --cafile string     the path to the file containing the ca certificate (chain) for the splunk HEC endpoint
      --hechost string    the hostname or IP of the splunk HEC endpoint
      --hecport int       port on which the splunk HEC endpoint is listening
  -h, --help              help for splunk-config-manifest
      --index string      the splunk index to use for this cluster's audit logs
      --tls               whether to use TLS encryption. You do need to specify a CA file.
      --token string      the hec token to use for this cluster's audit logs
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

* [cloudctl cluster](cloudctl_cluster.md)	 - manage clusters

