# SMTP Listener

The SMTP listener accepts email and sends them as cloudevents to Direktiv. The email content is in the `data` section of the cloud event. It containes `from`, `to`, `attachments` (Base64 encoded), `subject` and `message`. This listener can be installed in Knative and Plain mode. Plain mode is sending the events directly to Direktiv. Knative mode can use a broker and all other Knative Eventing features. Both modes can be installed with `kubectl apply`.

## Plain Mode

[plain.yaml](https://github.com/direktiv-listeners/smtp/blob/main/kubernetes/plain.yaml)

## Knative Mode

[knative.yaml](https://github.com/direktiv-listeners/smtp/blob/main/kubernetes/knative.yaml)

## Exposing TCP Services

The SMPT service needs to be exposed via a Kubernetes ingress controller. In case of NGINX the configuration in `direktiv.yaml` would look like the following code example.

```yaml
ingress-nginx:
  tcp:
    2525: default/smtp-listener-service:2525
```

## Configuration

| Environment Variable      | Description |
| ----------- | ----------- |
| DIREKTIV_SMTP_USERNAME      | Username authentication       |
| DIREKTIV_SMTP_PASSWORD   | Password for user authentication        |
| DIREKTIV_SMTP_ENDPOINT | Only required for the plain installation. The format is `http://direktiv-eventing.default/direktiv` where the last part is the target namespace |
| DIREKTIV_SMTP_TOKEN | Direktiv API key or access token |
| DIREKTIV_SMTP_INSEURE_TLS | If Direktiv uses a self-signed certifcate this needs to be set to `true` |
| DIREKTIV_SMTP_ADDRESS | The listener bind address |
| DIREKTIV_SMTP_HASH | If set to `true` the listener generates an cloud event ID based on the content of the email to avoid duplicate events. Random ID otherwise |
| DIREKTIV_SMTP_DEBUG | Enable debug logging |


