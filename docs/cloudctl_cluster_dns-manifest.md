## cloudctl cluster dns-manifest

create a manifest for an ingress or service type loadbalancer, creating a DNS entry and valid certificate within your cluster domain

```
cloudctl cluster dns-manifest <clusterid> [flags]
```

### Options

```
      --backend-name string    the name of the backend (default "my-backend")
      --backend-port int32     the port of the backend (default 443)
  -h, --help                   help for dns-manifest
      --ingress-class string   the ingress class name (default "nginx")
      --name string            the resource name (default "<name>")
      --namespace string       the resource's namespace (default "default")
      --ttl int                the ttl set to the created dns entry (default 180)
      --type string            either of type ingress or service (default "ingress")
      --with-certificate       whether to request a let's encrypt certificate for the requested dns entry or not (default true)
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

