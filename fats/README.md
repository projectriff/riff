# FaaS Acceptance Test Suite (FATS) for riff

![CI](https://github.com/projectriff/fats/workflows/CI/badge.svg)

FATS is a suite of scripts that support testing riff against various Kubernetes clusters.

## Running FATS

FATS is expected to be driven by the repository being tested.

An example config is provided in the `.github/workflows/ci.yaml` file for this repo that doubles as a test suite for FATS itself. The configuration defines environment variables and scripts to invoke.

- `CLUSTER_NAME` a short, url safe, unique name (e.g. fats-123-4) used to distinguish resource between concurrent runs.
- `NAMESPACE` the namespace to install resources into. May be hard coded for clusters that are provisioned on demand or dynamic for clusters that are reused between runs. (Note: if resources are shared jobs should not run concurrently)
- `CLUSTER` the type of cluster to use. Some clusters will require additional environment variables
- `REGISTRY` the type of registry to use. Some registries will require additional environment variables

There are several scripts that are commonly defined for FATS runs, environments may be configured differently.

- `start.sh` - starts the cluster and registry
- `.github/workflows/install.sh` - installs riff and dependencies into the cluster
- `.github/workflows/run.sh` - runs the tests against the cluster
- `diagnostics.sh` - dumps diagnostics information about the state of resources on the cluster and controller logs.
- `.github/workflows/cleanup.sh` - removes resources created by the tests run and removes riff
- `cleanup.sh` - shutting down the cluster and registry

Note: scripts in the `.github` directory are expected to be provided by each consumer of FATS, these scripts can be used as a starting point.

FATS will:

- configure and start a kubernetes cluster defined by `$CLUSTER`
- configure and start an image registry defined by `$REGISTRY`
- create, invoke (asserting correct output) and cleanup functions
- cleanup the cluster and registry after tests are complete
- provide Kubernetes Service type via `$K8S_SERVICE_TYPE`

You need to:

- pick the cluster (set as $CLUSTER, e.g. 'kind') and registry (set as $REGISTRY, e.g. 'dockerhub') to use, supplying any custom config they require.
- start FATS, typically:
  - `source ./start.sh`
- install and configure riff
- create and configure the target namespace, typically:
  - `kubectl create namespace $NAMESPACE`
  - `fats_create_push_credentials $NAMESPACE`
- create functions, applications and deployers to test:
  - `riff function create ${function_name} ...`
  - `riff ${runtime} deployer create ${deployer_name} --function-ref ${function_name}`
  - invoke
    - incluster `source ./macros/invoker-incluster.sh ${url} "${curl_opts}" ${expected_value}`
    - ingress `source ./macros/invoker-contour.sh ${url} "${curl_opts}" ${expected_value}`
- cleanup riff
- cleanup FATS, typically:
  - `source ./cleanup.sh`


## Extending FATS

There are four extension points for FATS:

- clusters: kubernetes clusters
- registries: image registries where built functions are pushed before they are pulled into the cluster
- functions: sample functions that can be invoked with helper scripts to aid creating, invoking and cleaning up
- tools: items that need to be installed, like kubectl or gcloud

### Clusters

Support is provided for:

- `kind`
  - Required credentials:
    - *none*
- `gke`
  - Required credentials:
    - `GCLOUD_CLIENT_SECRET`: base64 encoded json GCP service account token
- `pks-gcp`
  - Required credentials:
    - `GCLOUD_CLIENT_SECRET`: base64 encoded json GCP service account token
    - `TOOLSMITH_ENV`: base64 encoded toolsmiths credential file
    - `PIVNET_REFRESH_TOKEN` PivNet refresh token able to download the PKS CLI

To add a new cluster, create a directory under `./clusters/` and add three files:

- `configure.sh` - configuration shared by the start and cleanup scripts
  - do any other one time configuration for the cluster
- `start.sh` - start the kubernetes cluster and set it as the default kubectl context
- `cleanup.sh` - shutdown the running cluster and clean up any shared or external resources

### Registries

Support is provided for:

- `docker-daemon`
  - Credentials:
    - *none*
- `dockerhub`
  - Credentials:
    - `DOCKER_USERNAME` DockerHub username
    - `DOCKER_PASSWORD` DockerHub password
- `gcr`
  - Required credentials:
    - `GCLOUD_CLIENT_SECRET`: base64 encoded json GCP service account token

To add a new registry, create a directory under `./registries/` and add three files:

- `configure.sh` - configuration for the registry
  - define function `fats_image_repo` that echos the Docker repository to push the image to
  - define function `fats_delete_image` that deletes a published image
  - define function `fats_create_push_credentials` that creates a secret to be used to push in cluster builds to the registry
  - do any other one time configuration for the registry (run before the cluster is started)
- `start.sh` - start the registry and set it as the default for docker push (run after the cluster is started)
- `cleanup.sh` - shutdown the running registry and clean up any shared or external resources

### Functions

Support is provided for:

- uppercase
  - command
  - java
  - java-boot
  - node
  - npm

To add a new function, create a directory, adding the following files:

- `riff.toml` - containing values for `artifact` and `handler`, if needed
- files for your function

### Applications

Support is provided for:

- uppercase
  - java-boot
  - node

To add a new application, create a directory, adding the following files:

- files for your application

### Tools

Support is provided for:

- aws
- duffle
- glcoud
- helm
- kapp
- kind
- ko
- kubectl
- pivnet
- pks
- riff
- ytt

To add a new tool, create a file under `./tools/` as `<toolname>.sh`. Add any logic needed to install and configure the tool.
