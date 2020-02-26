![](https://github.com/projectriff/charts/workflows/CI/badge.svg)

# projectriff Release YAML

Release YAML files for riff (includes cert-manager, Knative, KEDA, kpack, Contour and dependencies).

## Install (kapp)

### Prerequisites

- a running kubernetes cluster (1.14+)
- [kubectl](https://kubectl.docs.kubernetes.io) (1.14+)
- [kapp](https://get-kapp.io) (0.14+)
- [ytt](https://get-ytt.io) (0.14+)

### Steps

1. Define riff version

   ```sh
   riff_version=0.5.0-snapshot

   kubectl create ns apps
   ```

1. Install riff Build (and dependencies)
   
   ```sh
   kapp deploy -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_version}/cert-manager.yaml
   ```

   ```sh
   kapp deploy -n apps -a kpack -f https://storage.googleapis.com/projectriff/release/${riff_version}/kpack.yaml
   ```

   ```sh
   kapp deploy -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-builders.yaml
   ```

   ```sh
   kapp deploy -n apps -a riff-build -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-build.yaml
   ```

1. Install Contour (for core or knative runtimes)
   
   If your cluster supports LoadBalancer services (most managed clusters do, but local clusters typically do not):

   ```sh
   kapp deploy -n apps -a contour -f https://storage.googleapis.com/projectriff/release/${riff_version}/contour.yaml
   ```
   
   If your cluster does not support LoadBalancer services, or if the above command stalls waiting for the ingress service to become ready, then you'll need to convert the ingress service to a NodePort:
   
   ```sh
   ytt -f https://storage.googleapis.com/projectriff/release/${riff_version}/contour.yaml -f https://storage.googleapis.com/projectriff/charts/overlays/service-nodeport.yaml --file-mark contour.yaml:type=yaml-plain | kapp deploy -n apps -a contour -f - -y
   ```

1. Optionally Install riff Core Runtime
   
   ```sh
   kapp deploy -n apps -a riff-core-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-core-runtime.yaml
   ```

1. Optionally Install riff Knative Runtime (and dependencies)
   
   ```sh
   kapp deploy -n apps -a knative -f https://storage.googleapis.com/projectriff/release/${riff_version}/knative.yaml
   ```

   ```sh
   kapp deploy -n apps -a riff-knative-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-knative-runtime.yaml
   ```

1. Optionally Install riff Streaming Runtime (and dependencies)
   
   ```sh
   kapp deploy -n apps -a keda -f https://storage.googleapis.com/projectriff/release/${riff_version}/keda.yaml
   ```

   ```sh
   kapp deploy -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-streaming-runtime.yaml
   ```

1. Enjoy.

### Uninstall

1. Remove any riff resources

   ```sh
   kubectl delete riff --all-namespaces --all
   ```

1. Remove riff Streaming Runtime (if installed)

   ```sh
   kapp delete -n apps -a riff-streaming-runtime
   ```

   ```sh
   kapp delete -n apps -a keda
   ```

1. Remove riff Knative Runtime (if installed)

   ```sh
   kubectl delete knative --all-namespaces --all
   ```

   ```sh
   kapp delete -n apps -a riff-knative-runtime
   ```

   ```sh
   kapp delete -n apps -a knative
   ```

1. Remove riff Core Runtime (if installed)
   ```sh
   kapp delete -n apps -a riff-core-runtime
   ```

1. Remove Contour (if installed)

   ```sh
   kapp delete -n apps -a contour
   ```

1. Remove riff Build

   ```sh
   kapp delete -n apps -a riff-build
   ```

   ```sh
   kapp delete -n apps -a riff-builders
   ```

   ```sh
   kapp delete -n apps -a kpack
   ```

   ```sh
   kapp delete -n apps -a cert-manager
   ```

## Creating installation YAML

### Prerequisites

- internet access
- [helm](https://helm.sh) (2.13+)
- [kapp](https://get-kapp.io) (0.14+)
- [ytt](https://get-ytt.io) (0.14+)
- [yq](http://mikefarah.github.io/yq/)
- [gcloud](https://cloud.google.com/sdk/gcloud/) (for publishing)

### Steps

Optionally, update the source templates to the latest component builds.

```sh
make templates
```

Package locally placing the YAML in the `target` directory.

```sh
make package
```
