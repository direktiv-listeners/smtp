# Testing

## Prepare 

Theses tests cover unit tests and end-to-end tests. To prepare the cluster two additional TCP ports need to be exposed with NGINX. Port `2525` is for the knative smtp listener and `2626` is the plain listener without Knative eventing.

```yaml
ingress-nginx:
  tcp:
    2525: default/smtp-listener-service-knative:2525
    2526: default/smtp-listener-service-plain:2526
```

After configuring these TCP ports, apply the `knative-test.yaml` and `plain-test.yaml` files with `kubectl apply -f`.

## Test

There is a Makefile target to run unit tests called `test`. To run the end-to-end test there needs to be a Direktiv server available with eventing enabled. The test function `TestSendServerE2E` which runs tests against an existing instances, e.g during development. It requires a server and a port where the knative port is `2525` and plain `2526`. 

```
TEST_PORT=2525 TEST_SERVER=192.168.0.145 go test -v -run TestSendServerE2E  cmd/*.go
TEST_PORT=2526 TEST_SERVER=192.168.0.145 go test -v -run TestSendServerE2E  cmd/*.go
```

Between tests calling `go clean -testcache` is required to clean the test cache.

## Build

docker build -t localhost:5000/smtp . && docker push localhost:5000/smtp
