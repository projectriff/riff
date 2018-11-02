package core

var (
	Buildpacks = map[string]string{
		"java": "projectriff/buildpack",
	}
	Invokers = map[string]string{
		"jar":     "https://github.com/projectriff/java-function-invoker/raw/v0.1.1/java-invoker.yaml",
		"command": "https://github.com/projectriff/command-function-invoker/raw/v0.0.7/command-invoker.yaml",
		"node":    "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
	}
	Manifests = map[string]*Manifest{
		"latest": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/latest/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/latest/riff-build.yaml",
				"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.pre.3.yaml",
			},
		},
		"stable": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20181029-1d5c521/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/previous/v20181029-e9f5b24/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/previous/v20181029-1d5c521/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181029-642bfc1/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181029-642bfc1/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
				"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.yaml",
			},
		},
		"v0.1.3": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20180921-69811e7/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20180921-69811e7/release-no-mon.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180921-01f95cb/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180921-01f95cb/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
				"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.yaml",
			},
		},
		"v0.1.2": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/release-no-mon.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
			},
		},
		"v0.1.1": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-crds.yaml",
				"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-main.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20180809-6b01d8e/release-no-mon.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
			},
		},
		"v0.1.0": &Manifest{
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/riff-releases/istio-riff-0.1.0.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/riff-releases/release-no-mon-riff-0.1.0.yaml",
				"https://storage.googleapis.com/riff-releases/release-eventing-riff-0.1.0.yaml",
				"https://storage.googleapis.com/riff-releases/release-eventing-clusterbus-stub-riff-0.1.0.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
			},
		},
	}
)

