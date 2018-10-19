# riff-distro

The `riff-distro` CLI provides commands to create a distribution of riff.
It uses a separate binary so that regular riff CLI users are not exposed to commands they'll never use, but it shares code and manifests (like `stable`) with the riff CLI. 

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

Download a stable release manifest and the corresponding release yaml files.
```sh
riff-distro system download -m stable -o .
```

Write a list of all images found in the release yaml files into `image-manifest.yaml`.  
Use ` --no-check` to avoid checking each image by pulling it down with your local docker daemon.
```sh
riff-distro image list -m stable
```

Download images and save them as files with sha256 names under `images`.  
Write the sha256 names into `image-manifest.yaml`.
```sh
riff-distro image pull -i image-manifest.yaml
```

Create the archive
```sh
tar czf ../riff-release.tar.gz .
```

## To install riff from an archive 

Download the archive, and then extract the files into an empty directory.
```sh
mkdir release-from-archive
cd release-from-archive
tar xzf ../riff-release.tar.gz
```

Relocate the images in the release into the `relocated` directory, specifying a new `dev.local` registry and `u` user, and rewriting all the release yaml with the relocated image names/tags. Using `--flatten` is required for registries which don't support deep hierarchies of image names.
```sh
riff image relocate -m manifest.yaml -r dev.local -u u -i image-manifest.yaml -o relocated --flatten
```

Load the relocated images from the file system into the registry.
```sh
cd relocated
riff image load -i image-manifest.yaml
```

Install riff and initialize the default namespace using relocated images.
```sh
riff system install -m manifest.yaml --node-port
riff namespace init default -m manifest.yaml --dockerhub $DOCKER_ID
```
