package img

import (
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Store interface {
	Ref() name.Reference
	Image() (v1.Image, error)
	Write(image v1.Image) error
}

func NewRegistry(ref string) (Store, error) {
	r, err := name.ParseReference(ref, name.WeakValidation)
	if err != nil {
		return nil, err
	}
	auth, err := authn.DefaultKeychain.Resolve(r.Context().Registry)
	if err != nil {
		return nil, err
	}
	return &registryStore{ref: r, auth: auth}, nil
}

type registryStore struct {
	ref  name.Reference
	auth authn.Authenticator
}

func (r *registryStore) Ref() name.Reference {
	return r.ref
}

func (r *registryStore) Image() (v1.Image, error) {
	return remote.Image(r.ref, remote.WithAuth(r.auth))
}

func (r *registryStore) Write(image v1.Image) error {
	return remote.Write(r.ref, image, r.auth, http.DefaultTransport, remote.WriteOptions{})
}

func NewDaemon(tag string) (Store, error) {
	t, err := name.NewTag(tag, name.WeakValidation)
	if err != nil {
		return nil, err
	}
	return &daemonStore{tag: t}, nil
}

type daemonStore struct {
	tag name.Tag
}

func (d *daemonStore) Ref() name.Reference {
	return d.tag
}

func (d *daemonStore) Image() (v1.Image, error) {
	return daemon.Image(d.tag)
}

func (d *daemonStore) Write(image v1.Image) error {
	_, err := daemon.Write(d.tag, image, daemon.WriteOptions{})
	return err
}
