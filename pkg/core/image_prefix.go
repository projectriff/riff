package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type RegistryOption struct {
	Value               string
	ImagePrefixSupplier func(string) (string, error)
}

func DockerRegistryOption(value string) *RegistryOption {
	return &RegistryOption{
		Value:               value,
		ImagePrefixSupplier: dockerHubImagePrefix,
	}
}

func GoogleContainerRegistryOption(value string) *RegistryOption {
	return &RegistryOption{
		Value:               value,
		ImagePrefixSupplier: gcrImagePrefix,
	}
}

func DetermineImagePrefix(userDefinedPrefix string, registryOptions ...*RegistryOption) (string, error) {
	if userDefinedPrefix != "" {
		return userDefinedPrefix, nil
	}
	for _, registryOption := range registryOptions {
		value := registryOption.Value
		if value != "" {
			return registryOption.ImagePrefixSupplier(value)
		}
	}
	return "", nil
}

func dockerHubImagePrefix(dockerHubId string) (string, error) {
	return fmt.Sprintf("docker.io/%s", dockerHubId), nil
}

func gcrImagePrefix(gcrTokenPath string) (string, error) {
	token, err := ioutil.ReadFile(gcrTokenPath)
	if err != nil {
		return "", err
	}
	tokenMap := map[string]string{}
	err = json.Unmarshal(token, &tokenMap)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gcr.io/%s", tokenMap["project_id"]), nil
}
