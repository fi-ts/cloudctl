# Gateway

## server

### Describe

```bash
# Describe gateway server
DEBUG=1 bin/cloudctl gateway server describe --name server
```

## Client

### Create

```bash
DEBUG=1 bin/cloudctl gateway client create --project=my-project --name=my-gw --pipes=nginx:8082:reverse-cluster-int-nginx:8082

# Change context to gateway client cluster's context (kind-kind2 in this case) by `kubectx`
kubectl ctx kind-kind2

# Change namespace to the gateway client instance's namespace (my-project in this case) by `kubens`
kubectl ns my-project

# Fetch the gateway client instance (my-gw in this case)
kubectl get instance my-gw -o yaml

# See the pipe in action
kubectl exec $(kubectl get pod -o jsonpath="{.items[0].metadata.name}") -- curl localhost:8082

# See the pipe in action via service
kubectl port-forward svc/cloudgateway-my-gw-nginx 8082:8082 &
curl localhost:8082

# See the pipe in action via `kubefwd`
sudo -E kubefwd svc
# In another terminal
curl cloudgateway-my-gw-nginx:8082

```

### Add Pipes

```bash
DEBUG=1 bin/cloudctl gateway client add-pipes --project=my-project --name=my-gw --pipes=echoserver:8088:reverse-echoserver.client:80

kubectl exec $(kubectl get pod -o jsonpath="{.items[0].metadata.name}") -- curl localhost:8088

kubectl port-forward svc/cloudgateway-my-gw-echoserver 8088:8088 &
curl localhost:8088

sudo -E kubefwd svc
# In another terminal
curl cloudgateway-my-gw-echoserver:8088
```

### Delete

```bash
DEBUG=1 bin/cloudctl gateway client delete --project=my-project --name=my-gw
```

### List

```bash
DEBUG=1 bin/cloudctl gateway client list
```

## Demo

```bash
bin/cloudctl gateway server create \
--project cloud-gateway-server server-a \
--ip 172.18.255.200 \
--pipes nginx:8080:reverse-cluster-int-nginx.default:8082 \
--pipes reverse:8088:reverse-echoserver.client:80 \
--pipes reverse-ext:8089:reverse-echoserver.client:80

bin/cloudctl gateway server describe --project cloud-gateway-server server-a

bin/cloudctl gateway client create \
--project cloud-gateway-client client-a \
--server cloud-gateway-server--server-a \
--pipes nginx

bin/cloudctl gateway client describe --project cloud-gateway-client client-a

tempo curl cloudgateway-client-a-nginx.cloud-gateway-client-a:8080 -m 5

bin/cloudctl gateway client patch -p cloud-gateway-client client-a --pipes-to-add=reverse

tempo curl cloudgateway-client-a-reverse.cloud-gateway-client:8088
```
