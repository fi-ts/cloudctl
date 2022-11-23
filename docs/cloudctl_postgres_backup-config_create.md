## cloudctl postgres backup-config create

create backup configuration

```
cloudctl postgres backup-config create [flags]
```

### Options

```
  -h, --help                      help for create
      --name string               name of the backup config
      --project string            project of the backup config
      --retention int32           number of backups per database to retain (default 10)
      --s3-accesskey string       s3-accesskey
      --s3-bucketname string      s3 bucketname to backup to
      --s3-encryptionkey string   s3 encryption key, enables sse (server side encryption) if given [optional]
      --s3-endpoint string        s3 endpoint to backup to
      --s3-region string          s3 region to backup to [optional]
      --s3-secretkey string       s3-secretkey
      --schedule string           backup schedule in cron syntax (default "30 00 * * *")
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

* [cloudctl postgres backup-config](cloudctl_postgres_backup-config.md)	 - manage postgres backup configuration

