# riff-distro

The `riff-distro` CLI provides commands to create a distribution of riff.

A distribution includes an install manifest, release yaml files, and image files.

`riff-distro` uses a separate binary so that regular riff CLI users are not exposed to commands they'll never use, but it shares code and manifests (like `stable`) with the riff CLI.

### To build riff-distro from source

Clone the riff project, and ensure that you have a working go environment with `dep`.

```sh
cd distro
make build   # build riff-distro
make install # copy to $GOPATH/bin
```

## To create a new distribution of riff

Start in an empty directory
```sh
mkdir ~/release-new
cd ~/release-new
```

Download an install manifest (e.g. `stable`) and the corresponding release yaml files.
```sh
riff-distro system download -m stable -o .
```

Scan the release yaml files to generate a list of images into `image-manifest.yaml`.  In general this list will require additional validation. 

```sh
riff-distro image list -m manifest.yaml --no-check
```

Using ` --no-check` avoids checking each image. Checks are performed by attempting to pull the images using the local docker daemon.

Download images and save them as files with sha256 names under `images`.  
Write the sha256 names into `image-manifest.yaml`.
```sh
riff-distro image pull -i image-manifest.yaml
```

Create the archive
```sh
tar czf ../riff-release.tar.gz .
```

## Install riff from a distribution

To install riff from a distribution, the images in the archive first have to be pushed into a registry, and a new install manifest created, pointing to release yaml with the "relocated" image names.

Download the archive, and then extract the files into an empty directory.
```sh
mkdir release-from-archive
cd release-from-archive
tar xzf ../riff-release.tar.gz
```

Relocate the images into a `relocated` directory, specifying the `registry` and `user-repo`. This will create a new manifest and new release yaml files, using `registry`/`user-repo` for each of the image names.

```sh
export REGISTRY_HOST=docker.io  # replace with your private registry
export REGISTRY_ID=???          # replace with your registry repo/account

riff image relocate \
  -m manifest.yaml \
  -r $REGISTRY_HOST \
  -u $REGISTRY_ID \
  -i image-manifest.yaml \
  -o relocated \
  --flatten
```

Using `--flatten` is required for registries which don't support deep hierarchies of image names like DockerHub.

Push the relocated images to a registry. This first loads the images from the file system and tags them in your local docker daemon.
The final `docker push` step requires docker push credentials for the registry.

```sh
riff image push -i relocated/image-manifest.yaml
```

To install riff locally (e.g. for minikube or Docker Desktop), use `riff image load` instead of `riff image push`.
This loads and tags images locally without pushing them to a registry.

Install riff and initialize the default namespace using the `relocated/manifest.yaml` and images.
```sh
riff system install -m relocated/manifest.yaml --node-port
riff namespace init default -m relocated/manifest.yaml --dockerhub $DOCKER_ID
```
The commands above assume that you are installing locally (--node-port) and initializing the namespace for `riff function create` to push to DockerHub.
