## cloudctl s3

manage s3 entities

### Synopsis

manages s3 users to access s3 storage located in different partitions.

### Options

```
  -h, --help   help for s3
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
* [cloudctl s3 add-key](cloudctl_s3_add-key.md)	 - adds a key for an s3 user
* [cloudctl s3 apply](cloudctl_s3_apply.md)	 - applies one or more s3 from a given file
* [cloudctl s3 client-config](cloudctl_s3_client-config.md)	 - returns fitting configuration of an s3 user for given client
* [cloudctl s3 create](cloudctl_s3_create.md)	 - creates the s3
* [cloudctl s3 delete](cloudctl_s3_delete.md)	 - deletes the s3
* [cloudctl s3 describe](cloudctl_s3_describe.md)	 - describes the s3
* [cloudctl s3 edit](cloudctl_s3_edit.md)	 - edit the s3 through an editor and update
* [cloudctl s3 list](cloudctl_s3_list.md)	 - list all s3
* [cloudctl s3 partitions](cloudctl_s3_partitions.md)	 - list s3 partitions
* [cloudctl s3 remove-key](cloudctl_s3_remove-key.md)	 - remove a key for an s3 user
* [cloudctl s3 update](cloudctl_s3_update.md)	 - updates the s3

