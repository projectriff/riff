module github.com/projectriff/riff

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/boz/go-lifecycle v0.0.0-20170921044039-c39961a5a0ce // indirect
	github.com/boz/go-logutil v0.0.0-20170814044541-9d21a9e4757d
	github.com/boz/kail v0.6.0
	github.com/boz/kcache v0.0.0-20171103002618-fb1338d32301
	github.com/buildpack/pack v0.0.9
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/frioux/shellquote v0.0.1
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/envy v1.6.10 // indirect
	github.com/golang/groupcache v0.0.0-20181024230925-c65c006176ff // indirect
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/uuid v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/knative/build v0.3.0
	github.com/knative/eventing v0.3.0
	github.com/knative/pkg v0.0.0-20190108184541-4365af623c75
	github.com/knative/serving v0.3.0
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/pkg/errors v0.8.0
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.2.2
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	golang.org/x/crypto v0.0.0-20181127143415-eb0de9b17e85
	google.golang.org/appengine v1.2.0 // indirect
	k8s.io/api v0.0.0-20180904230853-4e7be11eab3f
	k8s.io/apimachinery v0.0.0-20180904193909-def12e63c512
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be // indirect
)

// override is from pack@v0.0.9
replace github.com/google/go-containerregistry => github.com/dgodd/go-containerregistry v0.0.0-20180912122137-611aad063148a69435dccd3cf8475262c11814f6
