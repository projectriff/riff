# Samples for riff functions built on knative/serving project

The credentials for GCR need to be added to `credentials/gcr-credentials.yaml`.
Create a JSON key file using:

```bash
gcloud iam service-accounts keys create \
  --iam-account "push-images@cf-spring-funkytown.iam.gserviceaccount.com" \
  cf-spring-funkytown-push-images.json
```

Then replace the `{JSON key file content}` placeholder in `credentials/gcr-credentials.yaml`
with the content of the generated `cf-spring-funkytown-push-images.json` file.
Just make sure to keep it indented.

1. Install the buildtemplate for riff `templates/riff.yaml`
2. Create the secret for gcr `credentials/gcr-credentials.yaml`
3. Create the serviceaccount `serviceaccounts/riff-build.yaml`
4. Apply the functions - `function-node/service.yaml` or `function-java/service.yaml`

Look up the SERVICE_HOST and SERVICE_IP the same way as for other samples and curl the function.

For `function-node/service.yaml` you can use:

```bash
export SERVICE_HOST=`kubectl get route riff-square -o jsonpath="{.status.domain}"`
export SERVICE_IP=`kubectl get svc knative-ingressgateway -n istio-system -o jsonpath="{.status.loadBalancer.ingress[*].ip}"`
curl -w '\n' --header "Host:$SERVICE_HOST" --header "Content-Type: text/plain" http://${SERVICE_IP} -d 7
```

For `function-java/service.yaml` you can use:

```bash
export SERVICE_HOST=`kubectl get route riff-upper -o jsonpath="{.status.domain}"`
export SERVICE_IP=`kubectl get svc knative-ingressgateway -n istio-system -o jsonpath="{.status.loadBalancer.ingress[*].ip}"`
curl -w '\n' --header "Host:$SERVICE_HOST" --header "Content-Type: text/plain" http://${SERVICE_IP} -d knative
```
