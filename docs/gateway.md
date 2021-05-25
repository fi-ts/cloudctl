# Gateway

## Server

### Admin

```bash
# `add` adds the pipe to the server. If the server doesn't exist, create it.
# <pipes> == <pipe1,pipe2,...>, pipe1 == service-name:client-side-service-port:server-side-service-endpoint
$ cloudctl gw server add --project <project-uid> --name <server-name> --pipes <pipes>

# or use id
# <id> == <project-uid>--<server-name>
$ cloudctl gw server add <id> --pipes <pipes>

# `rm-pipes` removes the pipes from the server
$ cloudctl gw server rm-pipes <id> --pipes <pipes>

# `delete` deletes the server
$ cloudctl gw server delete <id>
```

### All Users

```bash
$ cloudctl gw server list
ID            Service           Pipe
<server-id-a> <service-name-a0> <pipe-a0>
              <service-name-a1> <pipe-a1>
<server-id-b> <service-name-b0> <pipe-b0>
              <service-name-b1> <pipe-b1>

$ cloudctl gw server describe <id>
Service           Pipe
<service-name-a0> <pipe-a0>
<service-name-a1> <pipe-a1>
```

## Client

```bash
# `add` enables the client to access the server's service. If the client doesn't exist, create it.
# <service-names> == service-name-1,service-name-2,...
$ cloudctl gw client create --project <project-uid> --name <client-name> --server-id <server-id> --services <service-names>

# or use id
# <id> == <project-uid>--<client-name>
$ cloudctl gw client create <id> --server-id <server-id> --services <service-names>

# `add-svc` adds the service to the client
$ cloudctl gw client patch <id> --service <service-name>

$ cloudctl gw client list
Name          Service           Endpoint
<client-id-a> <service-name-a0> <service-local-endpoint-a0>
              <service-name-a1> <service-local-endpoint-a1>
<client-id-b> <service-name-b0> <service-local-endpoint-b0>
              <service-name-b1> <service-local-endpoint-b1>

$ cloudctl gw client describe <id>
Service           Endpoint
<service-name-a0> <service-local-endpoint-a0>
<service-name-a1> <service-local-endpoint-a1>

# `rm-svc` removes the service from the client
$ cloudctl gw client rm-svc <id> --service <service-name>

# `delete` deletes the client
$ cloudctl gw client delete <id>
```
