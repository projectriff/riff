module github.com/projectriff/riff

go 1.12

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/boz/go-logutil v0.1.0
	github.com/boz/kail v0.10.1
	github.com/buildpack/pack v0.2.0
	github.com/fatih/color v1.7.0
	github.com/google/go-cmp v0.3.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/knative/pkg v0.0.0-20190510233738-9e0db8f0a7f4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/projectriff/system v0.0.0-20190530163417-0810dd8d9795
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190424203555-c05e17bb3b2d
	k8s.io/api v0.0.0-20190515023547-db5a9d1c40eb
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
	k8s.io/client-go v0.0.0-20190514184034-dd7f3ad83f18
)

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.10.2 // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/knative/build v0.6.0 // indirect
	github.com/knative/serving v0.6.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503 // indirect
)

replace (
	// kail wants to upgrade these deps, don't let it
	k8s.io/api => k8s.io/api v0.0.0-20190226173710-145d52631d00
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190226174127-78295b709ec6
)
