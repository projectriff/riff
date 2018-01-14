package minikube

import (
	"os/exec"
	"strings"
)

func QueryIp() (string, error) {
	cmdName := "minikube"

	cmd := exec.Command(cmdName, "ip")

	output, err := cmd.CombinedOutput()

	return strings.TrimRight(string(output),"\n"), err
}
