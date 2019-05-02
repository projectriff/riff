package core

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"syscall"
)

type SetCredentialsOptions struct {
	NamespaceName string

	SecretName   string
	GcrTokenPath string
	DockerHubId  string

	Registry     string
	RegistryUser string
}

func (o *SetCredentialsOptions) secretType() secretType {
	switch {
	case o.DockerHubId != "":
		return secretTypeDockerHub
	case o.GcrTokenPath != "":
		return secretTypeGcr
	case o.RegistryUser != "":
		return secretTypeBasicAuth
	default:
		return secretTypeUserProvided // should not happen...
	}
}

func (c *client) SetCredentials(options SetCredentialsOptions) error {
	namespace := options.NamespaceName
	secret, err := c.kubeClient.CoreV1().Secrets(namespace).Get(options.SecretName, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if secret != nil {
		fmt.Printf("Deleting existing secret %q in namespace %q\n", secret.ObjectMeta.Name, namespace)
		if err = c.kubeClient.CoreV1().Secrets(namespace).Delete(secret.Name, &v1.DeleteOptions{}); err != nil {
			return err
		}
	}
	initLabels := getInitLabels()
	secret, err = c.convertAdditionalSecret(&options, initLabels)
	if err = c.createBasicSecret(namespace, secret); err != nil {
		return err
	}

	serviceAccount, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Get(BuildServiceAccountName, v1.GetOptions{});
	if errors.IsNotFound(err) {
		return c.createServiceAccount(namespace, secret.Name, initLabels)
	}
	if err != nil {
		return err
	}
	return c.updateServiceAccount(namespace, secret.Name, serviceAccount)
}

func (c *client) convertDockerHubSecret(namespace, secret, username string, labels map[string]string) (*corev1.Secret, error) {
	return c.convertRegistrySecret(namespace, secret, username, "https://index.docker.io/v1/", labels)
}

func (c *client) convertGcrSecret(namespace, secret, gcrTokenPath string, labels map[string]string) (*corev1.Secret, error) {
	token, err := ioutil.ReadFile(gcrTokenPath)
	if err != nil {
		return nil, err
	}
	return c.convertToBasicAuthSecret(namespace, secret, "_json_key", string(token), "https://gcr.io", labels), nil
}

func (c *client) convertRegistrySecret(namespace, secret, username, registry string, labels map[string]string) (*corev1.Secret, error) {
	password, err := readPassword(fmt.Sprintf("Enter password for user %q", username))
	if err != nil {
		return nil, err
	}
	return c.convertToBasicAuthSecret(namespace, secret, username, password, registry, labels), nil
}

func (c *client) convertToBasicAuthSecret(namespace string,
	secretName string,
	username string,
	password string,
	serverAddress string,
	initLabels map[string]string) *corev1.Secret {

	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:        secretName,
			Annotations: map[string]string{"build.knative.dev/docker-0": serverAddress},
			Labels:      initLabels,
		},
		Type: corev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			"username": username,
			"password": password,
		},
	}
}

func (c *client) resetBasicSecret(namespace string, secret *corev1.Secret) error {
	_ = c.kubeClient.CoreV1().Secrets(namespace).Delete(secret.ObjectMeta.Name, &v1.DeleteOptions{})
	return c.createBasicSecret(namespace, secret)
}

func (c *client) createBasicSecret(namespace string, secret *corev1.Secret) error {
	fmt.Printf("Creating secret %q with basic authentication to server %q for user %q\n",
		secret.ObjectMeta.Name,
		secret.Annotations["build.knative.dev/docker-0"],
		secret.StringData["username"])

	_, err := c.kubeClient.CoreV1().Secrets(namespace).Create(secret)
	return err
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	if terminal.IsTerminal(int(syscall.Stdin)) {
		res, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		return string(res), err
	} else {
		reader := bufio.NewReader(os.Stdin)
		res, err := ioutil.ReadAll(reader)
		return string(res), err
	}
}

func (c *client) convertAdditionalSecret(options *SetCredentialsOptions, initLabels map[string]string) (*corev1.Secret, error) {
	namespace := options.NamespaceName
	secretName := options.SecretName
	switch options.secretType() {
	case secretTypeGcr:
		return c.convertGcrSecret(namespace, secretName, options.GcrTokenPath, initLabels)
	case secretTypeDockerHub:
		return c.convertDockerHubSecret(namespace, secretName, options.DockerHubId, initLabels)
	case secretTypeBasicAuth:
		return c.convertRegistrySecret(namespace, secretName, options.RegistryUser, options.Registry, initLabels)
	}
	return nil, nil
}

func (c *client) createServiceAccount(namespace, secretName string, initLabels map[string]string) error {
	serviceAccount := &corev1.ServiceAccount{}
	serviceAccount.Name = BuildServiceAccountName
	serviceAccount.Labels = initLabels
	serviceAccount.Secrets = updatedSecrets(corev1.ObjectReference{Name: secretName}, serviceAccount.Secrets)

	fmt.Printf("Creating serviceaccount %q with secret %q in namespace %q\n", serviceAccount.Name, secretName, namespace)
	_, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	return err
}

func (c *client) updateServiceAccount(namespace, secretName string, serviceAccount *corev1.ServiceAccount) error {
	serviceAccount.Secrets = updatedSecrets(corev1.ObjectReference{Name: secretName}, serviceAccount.Secrets)

	fmt.Printf("Adding secret %q to serviceaccount %q in namespace %q\n", secretName, serviceAccount.Name, namespace)
	_, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
	return err
}

func updatedSecrets(newSecret corev1.ObjectReference, boundSecrets []corev1.ObjectReference) []corev1.ObjectReference {
	var secrets []corev1.ObjectReference
	matched := false
	for _, boundSecret := range boundSecrets {
		secret := boundSecret
		if boundSecret.Name == newSecret.Name {
			matched = true
			secret = newSecret
		}
		secrets = append(secrets, secret)
	}
	if !matched {
		secrets = append(secrets, newSecret)
	}
	return secrets
}
