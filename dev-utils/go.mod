module github.com/projectriff/riff/dev-utils

go 1.13

require (
	github.com/projectriff/riff/stream-client-go v0.0.0
	github.com/spf13/cobra v0.0.6
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace github.com/projectriff/riff/stream-client-go => ../stream-client-go
