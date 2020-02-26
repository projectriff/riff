package authn

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gauthn "github.com/google/go-containerregistry/pkg/authn"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testResource struct {
	registry string
}

func (r *testResource) RegistryStr() string {
	return r.registry
}

func (r *testResource) String() string {
	return r.registry
}

func TestNewSecretsKeychain(t *testing.T) {
	tests := []struct {
		name     string
		secrets  []corev1.Secret
		resource gauthn.Resource
		expected gauthn.Authenticator
		wantErr  bool
	}{
		{
			name:    "empty",
			secrets: []corev1.Secret{},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: gauthn.Anonymous,
		},
		{
			name: "gcr secret",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							DockerSecretAnnotation: "https://gcr.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: &gauthn.Basic{
				Username: "gcr-user",
				Password: "gcr-pass",
			},
		},
		{
			name: "registry missing scheme",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"build.pivotal.io/docker": "gcr.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: &gauthn.Basic{
				Username: "gcr-user",
				Password: "gcr-pass",
			},
		},
		{
			name: "registry has trailing slash",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"build.pivotal.io/docker": "https://gcr.io/",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: &gauthn.Basic{
				Username: "gcr-user",
				Password: "gcr-pass",
			},
		},
		{
			name: "secret missing username",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-invalid-secret",
						Annotations: map[string]string{
							"build.pivotal.io/docker": "gcr.io",
						},
					},
					Data: map[string][]byte{
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			wantErr: true,
		},
		{
			name: "ignore secret without annotation",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-ignored-secret",
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: gauthn.Anonymous,
		},
		{
			name: "only include secrets with type basic-auth",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-ignored-secret",
						Annotations: map[string]string{
							"build.pivotal.io/docker": "gcr.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeDockercfg,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			expected: gauthn.Anonymous,
		},
		{
			name: "secret missing password",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-invalid-secret",
						Annotations: map[string]string{
							"build.pivotal.io/docker": "gcr.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "https://gcr.io",
			},
			wantErr: true,
		},
		{
			name: "multiple secrets",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"build.pivotal.io/docker": "https://gcr.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("gcr-user"),
						"password": []byte("gcr-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"build.pivotal.io/docker": "index.docker.io",
						},
					},
					Data: map[string][]byte{
						"username": []byte("docker-hub-user"),
						"password": []byte("docker-hub-pass"),
					},
					Type: corev1.SecretTypeBasicAuth,
				},
			},
			resource: &testResource{
				registry: "index.docker.io",
			},
			expected: &gauthn.Basic{
				Username: "docker-hub-user",
				Password: "docker-hub-pass",
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			actual, err := NewSecretsKeychain(c.secrets).Resolve(c.resource)
			if (err != nil) != c.wantErr {
				t.Errorf("NewSecretsKeychain() error = %v, wantErr %v", err, c.wantErr)
				return
			}
			if c.wantErr {
				return
			}
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("resolved authenticator (-expected, +actual) = %v", diff)
			}
		})
	}
}
