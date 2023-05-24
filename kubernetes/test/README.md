# Testing

## Prepare 

Theses tests cover unit tests and end-to-end tests. To prepare the cluster two additional TCP ports need to be exposed with NGINX. Port `2525` is for the knative smtp listener and `2526` is the plain listener without Knative eventing. The TLS test runs at `2527`.

```yaml
ingress-nginx:
  tcp:
    2525: default/smtp-listener-service-knative:2525
    2526: default/smtp-listener-service-plain:2526
    2527: default/smtp-listener-service-plain-tls:2527
```

After configuring these TCP ports, apply the `knative-test.yaml`, `plain-test.yaml`, `plain-test-tls.yaml` files with `kubectl apply -f`. The configuration within those files require a namespace `direktiv` to receive events. 

## Test

There is a Makefile target to run unit tests called `test`. To run the end-to-end test there needs to be a Direktiv server available with eventing enabled. The test function `TestSendServerE2E` which runs tests against an existing instances, e.g during development. It requires a server and a port where the knative port is `2525`, plain `2526` and plain-tls `2527`. 

```
TEST_PORT=2525 TEST_SERVER=192.168.0.145 go test -v -run TestSendServerE2E  cmd/*.go
TEST_PORT=2526 TEST_SERVER=192.168.0.145 go test -v -run TestSendServerE2E  cmd/*.go
TEST_PORT=2527 TEST_SERVER=192.168.0.145 go test -v -run TestSendServerE2ETLS  cmd/*.go
```

Between tests calling `go clean -testcache` is required to clean the test cache.

## Build

docker build -t localhost:5000/smtp . && docker push localhost:5000/smtp

## Filter

direktivctl events set-filter -n direktiv -a http://192.168.0.145 myfilter filter.js