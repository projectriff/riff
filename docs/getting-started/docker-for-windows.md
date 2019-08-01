---
id: docker-for-windows
title: Getting Started on Docker for Windows
sidebar_label: Docker for Windows
---

The following will help you get started running a riff function with Knative on Docker Community Edition for Windows.

### TL;DR

1. Install the latest release of Docker for Windows (Edge)
2. Configure the cluster and enable Kubernetes
3. Install Knative using the riff CLI
4. Create a function
5. Invoke the function

### install docker edge
Kubernetes and the kubectl CLI are now included with [Docker Desktop for Windows](https://docs.docker.com/docker-for-windows/install/). Docker Desktop for Windows requires [Hyper-V](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v#enable-the-hyper-v-role-through-settings) on Windows 10 Pro. We recommend downloading the latest Edge version which comes with a more recent release of Kubernetes.

![download Docker edge for mac](/img/docker-edge-for-windows-download.png)

### configure the VM
Once Docker is installed and running, open Settings by right-clicking the Docker tray icon and configure your VM with 4GB of memory and 4 CPUs in the Advanced settings tab. Click on Apply.

![configure Docker VM](/img/docker-vm-config-windows.png)

### allow sharing of the C: drive

In the Shared Drives settings, enable sharing for the C drive, and enter your Windows password when prompted. This will be used for persistent volume claims to provide cache storage during function builds.

![configure Docker VM](/img/docker-edge-windows-shared-drives.png)

### enable Kubernetes
Enable Kubernetes in the Kubernetes tab, click on Apply, and wait for the installation to complete and the cluster to start. If there is no Kubernetes tab, you may need to [switch to Linux containers](https://docs.docker.com/docker-for-windows/#switch-between-windows-and-linux-containers) first.

![enable Kubernetes](/img/docker-edge-kubernetes-windows.png)

If you previously had minikube or GKE configured, switch your kubectl context to "docker-desktop" using a PowerShell or command window.

```sh
kubectl config use-context docker-desktop
```

### monitor your cluster
At this point it is useful to monitor your Kubernetes cluster. This PowerShell function will call `kubectl get` every 3 seconds.

```powershell
function watchpods { while(1){ kubectl get pod --all-namespaces; start-sleep -seconds 3; clear }}
```

Start by watching all namespaces to confirm that Kubernetes is running.

```
watchpods
```
```
NAMESPACE     NAME                                     READY   STATUS    RESTARTS   AGE
docker        compose-7cf768cb84-nhl49                 1/1     Running   0          97s
docker        compose-api-579965d67f-j7nzj             1/1     Running   0          97s
kube-system   coredns-86c58d9df4-bxmr9                 1/1     Running   0          2m41s
kube-system   coredns-86c58d9df4-wqldh                 1/1     Running   0          2m41s
kube-system   etcd-docker-desktop                      1/1     Running   0          116s
kube-system   kube-apiserver-docker-desktop            1/1     Running   0          97s
kube-system   kube-controller-manager-docker-desktop   1/1     Running   0          108s
kube-system   kube-proxy-7n5v8                         1/1     Running   0          2m41s
kube-system   kube-scheduler-docker-desktop            1/1     Running   0          102s
```

### install the riff CLI
A zip with the riff CLI for Windows is available to download from our GitHub [releases](https://github.com/projectriff/riff/releases) page. Extract riff.exe and add it to a directory in your path. Once installed, validate the version.

```powershell
riff version
```
```
Version
  riff cli: 0.3.0 (4e474f57a463d4d2c1159af64d562532fcb3ac1b)
```


## install Knative using the riff CLI

Install Knative, watching the pods until everything is running (this could take a couple of minutes). The `--node-port` option replaces LoadBalancer type services with NodePort.

```powershell
riff system install --node-port
```

You should see pods running in namespaces istio-system, knative-build, knative-serving, and knative-eventing as well as kube-system when the system is fully operational. 

```
NAMESPACE          NAME                                            READY   STATUS      RESTARTS   AGE
docker             compose-7cf768cb84-nhl49                        1/1     Running     0          4m49s
docker             compose-api-579965d67f-j7nzj                    1/1     Running     0          4m49s
istio-system       cluster-local-gateway-547467ccf6-mpjh4          1/1     Running     0          2m24s
istio-system       istio-citadel-7d64db8bcf-hflr5                  1/1     Running     0          2m25s
istio-system       istio-cleanup-secrets-6s6vj                     0/1     Completed   0          2m35s
istio-system       istio-egressgateway-6ddf4c8bd6-h8v24            1/1     Running     0          2m25s
istio-system       istio-galley-7dd996474-54h46                    1/1     Running     0          2m25s
istio-system       istio-ingressgateway-84b89d647f-dz929           1/1     Running     0          2m25s
istio-system       istio-pilot-54b76645df-qpmr5                    2/2     Running     0          2m10s
istio-system       istio-policy-5c4d9ff96b-djrjn                   2/2     Running     0          2m25s
istio-system       istio-sidecar-injector-6977b5cf5b-9b8qm         1/1     Running     0          2m25s
istio-system       istio-statsd-prom-bridge-b44b96d7b-vlbk2        1/1     Running     0          2m25s
istio-system       istio-telemetry-7676df547f-g578v                2/2     Running     0          2m25s
knative-build      build-controller-7b8987d675-7pqrd               1/1     Running     0          99s
knative-build      build-webhook-74795c8696-mstpv                  1/1     Running     0          99s
knative-eventing   eventing-controller-864657d8d4-bvzl2            1/1     Running     0          94s
knative-eventing   in-memory-channel-controller-f794cc9d8-hkb4m    1/1     Running     0          92s
knative-eventing   in-memory-channel-dispatcher-8595c7f8d7-4s7cx   2/2     Running     2          92s
knative-eventing   webhook-5d76776d55-p95tf                        1/1     Running     0          94s
knative-serving    activator-bf6bffbc5-ntcv2                       2/2     Running     1          97s
knative-serving    autoscaler-86dfc64d87-jdhm9                     2/2     Running     1          97s
knative-serving    controller-b9c5d7fb8-tsqhq                      1/1     Running     0          96s
knative-serving    webhook-787c95f8bd-bh757                        1/1     Running     0          96s
kube-system        coredns-86c58d9df4-bxmr9                        1/1     Running     0          5m53s
kube-system        coredns-86c58d9df4-wqldh                        1/1     Running     0          5m53s
kube-system        etcd-docker-desktop                             1/1     Running     0          5m8s
kube-system        kube-apiserver-docker-desktop                   1/1     Running     0          4m49s
kube-system        kube-controller-manager-docker-desktop          1/1     Running     0          5m
kube-system        kube-proxy-7n5v8                                1/1     Running     0          5m53s
kube-system        kube-scheduler-docker-desktop                   1/1     Running     0          4m54s
```

## initialize the namespace and provide credentials for pushing images to DockerHub

Use the riff CLI in Windows PowerShell to initialize your namespace (if you plan on using a namespace other than `default` then substitute the name you want to use). This will create a serviceaccount and a secret with the provided credentials and install a buildtemplate. Replace the ??? with your docker username.

```powershell
riff namespace init default --docker-hub ???
```

You will be prompted to provide the password.

## create a function from a GitHub repo

This riff command (formatted for PowerShell) will pull the source code for a function from a GitHub repo, build a container image based on the node function invoker, and push the resulting image to your dockerhub repo.

```powershell
riff function create square `
  --git-repo https://github.com/projectriff-samples/node-square `
  --artifact square.js `
  --verbose
```

## invoke the function
```powershell
riff service invoke square --json -- -w '\n' -d 8
```

#### result
```
curl http://localhost:31380/ -H 'Host: square.default.example.com' -H 'Content-Type: application/json' -w '\n' -d 8
64
```

## create a function from code in a local directory

You can use riff to build functions from source in a local directory, instead of first committing the code to a repo on GitHub.

For this to work with Docker Hub from Windows, a small workaround is required to support the multiple ways the Docker Hub registry can be referenced.

After being prompted for you docker credentials you should see 2 entries for docker.io:

```powershell
$dockerid = Read-Host "Please enter your Docker ID: "
$password = Read-Host -assecurestring "Please enter your password: "
echo $('{"ServerURL": "https://index.docker.io", "Username": "' +  $dockerid + '", "Secret": "' + [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($password)) + '"}') | docker-credential-wincred store
docker-credential-wincred list
```

#### result
```
{"https://index.docker.io":"YOUR_USERNAME","https://index.docker.io/v1/":"YOUR_USERNAME"}
```

### create the function

Using PowerShell in a new directory with a single file called `square.js`

#### square.js
```js
module.exports = (x) => `the square of ${x} is ${x**2}`
```

#### delete the old square function and create a new one
```powershell
riff service delete square

riff function create square `
  --local-path . `
  --artifact square.js `
  --verbose
```

#### invoke the function
```powershell
riff service invoke square --json -- -w '\n' -d 8
```

#### result
```
curl http://localhost:31380/ -H 'Host: square.default.example.com' -H 'Content-Type: application/json' -w '\n' -d 8
the square of 8 is 64
```

### Note
Due to differences between Windows and Linux file permissions, **Command invoker** functions which depend on file permissions will not work when built locally on Windows. We suggest building from a git repository as a workaround.

## uninstalling and reinstalling
If you need to upgrade or reinstall riff, we recommend resetting the Kubernetes cluster first. To do this, click `Reset Kubernetes Cluster...` in the Reset tab in Docker Settings.

![reset Kubernetes](/img/docker-edge-kubernetes-reset-windows.png)
