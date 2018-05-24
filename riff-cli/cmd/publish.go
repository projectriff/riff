/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/juju/errgo/errors"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/minikube"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
	"strconv"
)

type publishOptions struct {
	contentType string
	input       string
	data        string
	reply       bool
	count       int
	pause       int
}

func Publish(kube kubectl.KubeCtl, minik minikube.Minikube) *cobra.Command {

	var publishOptions publishOptions

	// publishCmd represents the publish command
	var publishCmd = &cobra.Command{
		Use:   "publish",
		Short: "Publish data to a topic using the http-gateway",
		Long:  `Publish data to a topic using the http-gateway`,
		Example: `
	riff publish -i greetings -d hello -r
	
will post 'hello' to the 'greetings' topic and wait for a reply.

	riff publish --content-type application/json -i concat -r -d '{"hello":"world"}'

will post '{"hello":"world"}' as json to the 'concat' topic and wait for a reply.

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ipAddress, port, err := lookupAddress(kube, minik)
			if err != nil {
				return err
			}
			return publish(ipAddress, port, publishOptions)

		},
	}
	publishCmd.Flags().StringVarP(&publishOptions.data, "data", "d", "", "the data to post to the http-gateway using the input topic")
	publishCmd.Flags().StringVarP(&publishOptions.input, "input", "i", osutils.GetCWDBasePath(), "the name of the input topic, defaults to name of current directory")
	publishCmd.Flags().BoolVarP(&publishOptions.reply, "reply", "r", false, "wait for a reply containing the results of the function execution")
	publishCmd.Flags().IntVarP(&publishOptions.count, "count", "c", 1, "the number of times to post the data")
	publishCmd.Flags().IntVarP(&publishOptions.pause, "pause", "p", 0, "the number of seconds to wait between postings")
	publishCmd.Flags().StringVarP(&publishOptions.contentType, "content-type", "", "text/plain", "the content type")
	publishCmd.Flags().String("namespace", "", "the namespace of the http-gateway")

	publishCmd.MarkFlagRequired("data")
	publishCmd.Flags().MarkDeprecated("namespace", "it will be removed in future releases")

	return publishCmd
}

func lookupAddress(kube kubectl.KubeCtl, minik minikube.Minikube) (string, string, error) {
	cmdArgs := []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}
	output, err := kube.Exec(cmdArgs)

	if err != nil {
		return "", "", fmt.Errorf("Error querying http-gateway %v\n %v", err, output)
	}

	parser := jsonpath.NewParser([]byte(output))

	portType, err := parser.StringValue(`$.items[0].spec.type`)

	if err != nil {
		return "", "", errors.New("Unable to locate http-gateway: " + err.Error())
	}

	var ipOrHostname string
	var pFloat interface{}

	switch portType {
	case "NodePort":
		ipOrHostname, err = minik.QueryIp()
		if err != nil || strings.Contains(ipOrHostname, "Error getting IP") {
			ipOrHostname = "127.0.0.1"
		}
		pFloat, err = parser.Value(`$.items[0].spec.ports[?(@.name == http)].nodePort[0]`)
	case "LoadBalancer":
		ipOrHostname, err = parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].ip`)
		if ipOrHostname == "" {
			ipOrHostname, err = parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].hostname`)
			if ipOrHostname == "" {
				return "", "", errors.New("unable to determine http-gateway ip address nor hostname")
			}
		}
		pFloat, err = parser.Value(`$.items[0].spec.ports[?(@.name == http)].port[0]`)

	default:
		return "", "", fmt.Errorf("Unkown port type %s", portType)
	}

	if err != nil {
		return "", "", errors.New("Unable to determine gateway port: " + err.Error())
	}
	port := strconv.FormatFloat(pFloat.(float64), 'f', 0, 64)

	return ipOrHostname, port, nil
}

func publish(ipAddress string, port string, publishOptions publishOptions) error {
	resource := "messages"
	if publishOptions.reply {
		resource = "requests"
	}

	url := fmt.Sprintf("http://%s:%s/%s/%s", ipAddress, port, resource, publishOptions.input)

	fmt.Printf("Posting to %s\n", url)

	for i := 0; i < publishOptions.count; i++ {

		if result, err := doPost(url, publishOptions); err != nil {
			return err
		} else {
			fmt.Println(result)
		}

		if publishOptions.pause > 0 {
			time.Sleep(time.Duration(publishOptions.pause) * time.Second)
		}
	}
	return nil
}

func doPost(url string, publishOptions publishOptions) (string, error) {
	resp, err := http.Post(url, publishOptions.contentType, strings.NewReader(publishOptions.data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if 200 <= resp.StatusCode && resp.StatusCode < 400 {
		return string(body), nil
	} else {
		message := string(body)
		if resp.StatusCode == 504 {
			message = "Gateway Timeout"
		}
		return "", fmt.Errorf("HTTP Gateway responded with code %v: %v", resp.StatusCode, message)
	}

}
