package authn

import (
	"fmt"
	"strings"

	ggcrauthn "github.com/google/go-containerregistry/pkg/authn"
	corev1 "k8s.io/api/core/v1"
)

type DockerSecretsKeychain struct {
	regAuths map[string]*ggcrauthn.Basic
}

const DockerSecretAnnotation = "build.pivotal.io/docker"

func NewSecretsKeychain(secrets []corev1.Secret) ggcrauthn.Keychain {
	k := &DockerSecretsKeychain{
		regAuths: map[string]*ggcrauthn.Basic{},
	}
	for _, secret := range secrets {
		if secret.Annotations[DockerSecretAnnotation] == "" {
			continue
		}
		if secret.Type != corev1.SecretTypeBasicAuth {
			continue
		}
		k.regAuths[trimReg(secret.Annotations[DockerSecretAnnotation])] = &ggcrauthn.Basic{
			Username: string(secret.Data[corev1.BasicAuthUsernameKey]),
			Password: string(secret.Data[corev1.BasicAuthPasswordKey]),
		}
	}
	return k
}

func (k *DockerSecretsKeychain) Resolve(resource ggcrauthn.Resource) (ggcrauthn.Authenticator, error) {
	for reg, auth := range k.regAuths {
		if reg == trimReg(resource.RegistryStr()) {
			if auth.Username == "" {
				return nil, fmt.Errorf("invalid auth: missing username")
			}
			if auth.Password == "" {
				return nil, fmt.Errorf("invalid auth: missing password")
			}
			return auth, nil
		}
	}
	return ggcrauthn.Anonymous, nil
}

func trimReg(reg string) string {
	reg = strings.TrimPrefix(reg, "http://")
	reg = strings.TrimPrefix(reg, "https://")
	reg = strings.TrimRight(reg, "/")
	return reg
}
