## cloudctl postgres

manage postgres

### Synopsis


Create and manage postgres databases.

To create a postgres database you first need to specify/create a backup-config where the database backups are stored in.
This can be done either by auto-creating an S3 endpoint or provide all S3 details manually.

Example Workflow:

1. Display all possible locations where postgres databases can be created:

# cloudctl postgres partition
NAME            ALLOWEDTENANTS
dc1             any

2. Create a backup-config with retention count and schedule

# cloudctl postgres backup-config auto-create --name daily-for-one-week --project <your-project-id> --partition dc1 --retention 7 --schedule "45 3 * * 0"
ID                                      NAME                    PROJECT                                 SCHEDULE        RETENTION       S3                                                              CREATEDBY
3094421c-ee11-4155-b4d9-7fdac116c0ff    daily-for-one-week      b621eb99-4888-4911-93fc-95854fc030e8    45 3 * * *       7               https://s3.dev.example/backup-3094421c      <Achim Muster>[achim.muster@example.com]

3. Create a postgres database

# cloudctl postgres create --description accounting-db-test --project <your-project-id> --partition dc1 --backup-config <backup-config-id-from-above>
ID                                      DESCRIPTION             PARTITION       TENANT  PROJECT                                 CPU     BUFFER  STORAGE BACKUP-CONFIG                           REPLICAS VERSION AGE     STATUS
890b1601-6cc3-46cd-86a6-d4479bc1528d    accounting-db-test      dc1             fits    b621eb99-4888-4911-93fc-95854fc030e8    500m    500m    10Gi    3094421c-ee11-4155-b4d9-7fdac116c0ff    1        12      0s

4. Check if it is running with

# cloudctl postgres ls --description accounting-db-test
ID                                      DESCRIPTION             PARTITION       TENANT  PROJECT                                 CPU     BUFFER  STORAGE BACKUP-CONFIG                           REPLICAS VERSION AGE     STATUS
890b1601-6cc3-46cd-86a6-d4479bc1528d    accounting-db-test      dc1             fits    b621eb99-4888-4911-93fc-95854fc030e8    500m    500m    10Gi    3094421c-ee11-4155-b4d9-7fdac116c0ff    1        12      1m 21s  Running

5. Connect to the database

# cloudctl postgres connectionstring 890b1601-6cc3-46cd-86a6-d4479bc1528d
PGPASSWORD=J34JnhbtPQ2s1znPmp1pWNRv9EPvbsUvQ3OLh2ycyVAcMmK6upazlzM4JAELpaC0 psql --host=1.2.3.4 --port=32004 --username=postgres
PGPASSWORD=imQw7KYOdha2wLODlQFFEodN8eyfoFYKOUmXnxFwQJHIeOQtjZJmMGDskpa3SARr psql --host=1.2.3.4 --port=32004 --username=standby

# PGPASSWORD=J34JnhbtPQ2s1znPmp1pWNRv9EPvbsUvQ3OLh2ycyVAcMmK6upazlzM4JAELpaC0 psql --host=1.2.3.4 --port=32004 --username=postgres
psql (12.6 (Ubuntu 12.6-0ubuntu0.20.10.1))
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

postgres=#

6. You can create more databases, all using the same backup-config


### Options

```
  -h, --help   help for postgres
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
* [cloudctl postgres apply](cloudctl_postgres_apply.md)	 - apply postgres
* [cloudctl postgres backup-config](cloudctl_postgres_backup-config.md)	 - manage postgres backup configuration
* [cloudctl postgres connectionstring](cloudctl_postgres_connectionstring.md)	 - return the connectionstring for a postgres
* [cloudctl postgres create](cloudctl_postgres_create.md)	 - create postgres
* [cloudctl postgres create-standby](cloudctl_postgres_create-standby.md)	 - create postgres standby
* [cloudctl postgres delete](cloudctl_postgres_delete.md)	 - delete a postgres
* [cloudctl postgres demote-to-standby](cloudctl_postgres_demote-to-standby.md)	 - demote a the replication primary to become a standby instance
* [cloudctl postgres describe](cloudctl_postgres_describe.md)	 - describe a postgres
* [cloudctl postgres edit](cloudctl_postgres_edit.md)	 - edit postgres
* [cloudctl postgres list](cloudctl_postgres_list.md)	 - list postgres
* [cloudctl postgres list-backups](cloudctl_postgres_list-backups.md)	 - list postgres backups
* [cloudctl postgres partition](cloudctl_postgres_partition.md)	 - describe all partitions where postgres might be deployed
* [cloudctl postgres promote-to-primary](cloudctl_postgres_promote-to-primary.md)	 - promote a standby instance to become a replication primary
* [cloudctl postgres restore](cloudctl_postgres_restore.md)	 - restore postgres from existing one
* [cloudctl postgres restore-accepted](cloudctl_postgres_restore-accepted.md)	 - confirm the restore of a database
* [cloudctl postgres version](cloudctl_postgres_version.md)	 - describe all postgres versions

