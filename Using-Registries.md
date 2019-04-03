## Using Registries

### Using Docker Hub with builds

#### Initialize a namespace

When you initialize the namespace use the `--docker-hub` flag together with your Docker ID.

If you are initializing the `default` namespace with Docker Hub, then you can use:

```
riff namespace init default --docker-hub <docker-id>
```

### Using Google Container Registry (GCR) with builds

Make sure that you have installed the most recent version of the [Cloud SDK](https://cloud.google.com/sdk/docs/).

#### Create a Kubernetes secret for pushing images to GCR and initialize a namespace

1. Create a service account and assign it `roles/storage.admin` for your project's GCR bucket

        export GCP_PROJECT=$(gcloud config get-value core/project)
        gcloud iam service-accounts create push-image
        gcloud projects add-iam-policy-binding $GCP_PROJECT \
          --member serviceAccount:push-image@$GCP_PROJECT.iam.gserviceaccount.com \
          --role roles/storage.admin

2. Create and download the JSON key for the service account

        gcloud iam service-accounts keys create \
          --iam-account "push-image@$GCP_PROJECT.iam.gserviceaccount.com" \
          $HOME/push-image.json

    > NOTE: Store this JSON key file in a safe place for future use

3. Initialize a namespace

    When you initialize the namespace use the `--gcr` flag.
    If you are initializing the `default` namespace with GCR, then you can use:

        riff namespace init default --gcr ~/path/to/push-image.json


### Authenticating with Docker for local builds

First follow the instructions above for initializing a namespace with your registry credentials.

For local builds (using `--local-path` flag) then make sure your have authenticated Docker for the registry you are using.

* Docker Hub

        docker login

* GCR

    Make sure that you have installed the most recent version of the [Cloud SDK](https://cloud.google.com/sdk/docs/), which includes the `gcloud` command-line tool.
    See [Container Registry - Authentication methods](https://cloud.google.com/container-registry/docs/advanced-authentication) for more details.

    To use GCR for images created by local builds, configure Docker using:

        gcloud auth configure-docker


### Using private GCR images

When you create a function or deploy a service the image for the application/function will be pulled from the registry specified in the `--image` flag. In order to use a private registry for this image you need to provide a secret containing your image pull credentials and add it to `imagePullSecrets` for the default service account in the namespace you are using for the application/function. This secret is used when the image for your function or service is pulled during pod initialization.

> NOTE: The `namespace init` command creates a separate secret that is only used for function builds and subsequent push to the repository. This build secret is not used when pulling the image for running the service.

For example, when deploying images from a private Google Container Registry to the default namespace, you could do the following:

1. Create a service account and assign it `roles/storage.objectViewer` for your project's GCR bucket

        export GCP_PROJECT=$(gcloud config get-value core/project)
        gcloud iam service-accounts create pull-image
        gcloud projects add-iam-policy-binding $GCP_PROJECT \
          --member serviceAccount:pull-image@$GCP_PROJECT.iam.gserviceaccount.com \
          --role roles/storage.objectViewer

2. Create and download the JSON key for the service account

        gcloud iam service-accounts keys create \
          --iam-account "pull-image@$GCP_PROJECT.iam.gserviceaccount.com" \
          $HOME/$GCP_PROJECT-pull-image.json

    > NOTE: Store this JSON key file in a safe place for future use

3. Use the JSON key file to create the Secret

    > NOTE: If your moved the JSON key file to a different location make sure to use that location when setting the PULL_IMAGE_JSON_KEY environment variable below.

        export PULL_IMAGE_JSON_KEY=$HOME/$GCP_PROJECT-pull-image.json
        kubectl create secret docker-registry "gcr" \
          --docker-server=gcr.io \
          --docker-username=_json_key \
          --docker-password="$(cat $PULL_IMAGE_JSON_KEY)" \
          --docker-email=$(gcloud config get-value core/account) \
          --namespace "default"

4. Patch the default service account in the namespace which in this example is the `default` namespace

    > NOTE: If your service account already has an imagePullSecret specified you should add that secret to the array of secrets in the following command since the patch will replace the array.

        kubectl patch serviceaccount "default" \
          --patch '{"imagePullSecrets": [{"name": "gcr"}]}' \
          --namespace "default"
