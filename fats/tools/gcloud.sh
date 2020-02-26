#!/bin/bash

if ! [ -x "$(command -v gcloud)" ]; then
  if hash choco 2>/dev/null; then
    choco install gcloudsdk --ignore-checksums

    # expose gcloud to the path
    cat <<EOF > /usr/bin/gcloud
#!/bin/bash

"/c/Program Files (x86)/Google/Cloud SDK/google-cloud-sdk/bin/gcloud" \$@
EOF

    # expose all gcloudsdk *.cmd into the path
    while read -r cmd; do
      cat <<EOF > /usr/bin/${cmd}
#!/bin/bash

"/c/Program Files (x86)/Google/Cloud SDK/google-cloud-sdk/bin/${cmd}.cmd" \$@
EOF
    done <<< "$(ls -1 '/c/Program Files (x86)/Google/Cloud SDK/google-cloud-sdk/bin/' | grep \.cmd | cut -d. -f1)"
  else
    # Create environment variable for correct distribution
    export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"

    # Add the Cloud SDK distribution URI as a package source
    echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

    # Import the Google Cloud Platform public key
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

    # Update the package list and install the Cloud SDK
    sudo apt-get update && sudo apt-get install google-cloud-sdk
  fi
fi

gcloud config set project cf-spring-pfs-eng
gcloud config set compute/region us-central1
gcloud config set compute/zone us-central1-a
gcloud config set disable_prompts True

echo $GCLOUD_CLIENT_SECRET | base64 --decode > key.json
gcloud auth activate-service-account --key-file key.json
rm key.json
