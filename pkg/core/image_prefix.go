package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func DetermineImagePrefix(userDefinedPrefix string, dockerHubId string, gcrTokenPath string) (string, error) {
	if userDefinedPrefix != "" {
		return userDefinedPrefix, nil
	}
	if dockerHubId != "" {
		return dockerHubImagePrefix(dockerHubId), nil
	}
	if gcrTokenPath != "" {
		return gcrImagePrefix(gcrTokenPath)
	}
	return "", nil
}

func dockerHubImagePrefix(dockerHubId string) string {
	return fmt.Sprintf("docker.io/%s", dockerHubId)
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
