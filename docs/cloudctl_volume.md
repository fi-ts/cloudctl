## cloudctl volume

manage volume entities

### Synopsis

manage persistent cloud storage volumes

### Options

```
  -h, --help   help for volume
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
* [cloudctl volume clusterinfo](cloudctl_volume_clusterinfo.md)	 - show storage cluster infos
* [cloudctl volume delete](cloudctl_volume_delete.md)	 - deletes the volume
* [cloudctl volume describe](cloudctl_volume_describe.md)	 - describes the volume
* [cloudctl volume encryption-secret-manifest](cloudctl_volume_encryption-secret-manifest.md)	 - print a secret manifest for volume encryption
* [cloudctl volume list](cloudctl_volume_list.md)	 - list all volumes
* [cloudctl volume manifest](cloudctl_volume_manifest.md)	 - print a manifest for a volume
* [cloudctl volume snapshot](cloudctl_volume_snapshot.md)	 - manage snapshot entities

