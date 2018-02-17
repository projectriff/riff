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
	"strings"
	"fmt"
	"net/http"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/minikube"
	"github.com/projectriff/riff-cli/pkg/jsonpath"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/spf13/viper"
)

type PublishOptions struct {
	namespace string
	input string
	data  string
	reply bool
	count int
	pause int
}

var publishOptions PublishOptions

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish data to a topic using the http-gateway",
	Long: `Publish data to a topic using the http-gateway. For example:

    riff publish -i greetings -d hello -r

will post 'hello' to the 'greetings' topic and wait for a reply.
`,
	Run: func(cmd *cobra.Command, args []string) {

		// get the viper value from env var, config file or flag option
		publishOptions.namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

		// look for PUBLISH_NAMESPACE
		if !cmd.Flags().Changed("namespace") && viper.GetString("PUBLISH_NAMESPACE") != "" {
			publishOptions.namespace = viper.GetString("PUBLISH_NAMESPACE")
			fmt.Printf("Using namespace: %s\n", publishOptions.namespace)
		} else {
			// look for publishNamespace
			if !cmd.Flags().Changed("namespace") && viper.GetString("publishNamespace") != "" {
				publishOptions.namespace = viper.GetString("publishNamespace")
				fmt.Printf("Using namespace: %s\n", publishOptions.namespace)
			}
		}

		cmdArgs := []string{"get", "--namespace", publishOptions.namespace, "svc", "-l", "component=http-gateway", "-o", "json"}
		output, err := kubectl.ExecForBytes(cmdArgs)

		if err != nil {
			ioutils.Errorf("Error querying http-gateway %v\n %v\n", err, output)
			return
		}

		parser := jsonpath.NewParser(output)

		portType := parser.Value(`$.items[0].spec.type+`)

		if portType == "" {
			ioutils.Errorf("Unable to locate http-gateway in namespace %v\n", publishOptions.namespace)
			return
		}

		var ipAddress string
		var port string

		switch portType {
		case "NodePort":
			ipAddress, err = minikube.QueryIp()
			if err != nil || strings.Contains(ipAddress, "Error getting IP") {
				ipAddress = "127.0.0.1"
			}
			port = parser.Value(`$.items[0].spec.ports[*]?(@.name == "http").nodePort+`)
		case "LoadBalancer":
			ipAddress = parser.Value(`$.items[0].status.loadBalancer.ingress[0].ip+`)
			if ipAddress == "" {
				ioutils.Error("unable to determine http-gateway ip address")
				return
			}
			port = parser.Value(`$.items[0].spec.ports[*]?(@.name == "http").port+`)

		default:
			ioutils.Errorf("Unkown port type %s", portType)
			return
		}

		if port == "" {
			ioutils.Error("Unable to determine gateway port")
			return
		}

		publish(ipAddress, port)

	},
}

func publish(ipAddress string, port string) {
	resource := "messages"
	if publishOptions.reply {
		resource = "requests"
	}

	url := fmt.Sprintf("http://%s:%s/%s/%s", ipAddress, port, resource, publishOptions.input)

	fmt.Printf("Posting to %s\n", url)

	for i := 0; i < publishOptions.count; i++ {

		resp, err := http.Post(url, "text/plain", strings.NewReader(publishOptions.data))
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))

		if (publishOptions.pause > 0) {
			time.Sleep(time.Duration(publishOptions.pause) * time.Second)
		}
	}
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	publishCmd.Flags().StringVarP(&publishOptions.data, "data", "d", "", "the data to post to the http-gateway using the input topic")
	publishCmd.Flags().StringVarP(&publishOptions.input, "input", "i", osutils.GetCWDBasePath(), "the name of the input topic, defaults to name of current directory")
	publishCmd.Flags().BoolVarP(&publishOptions.reply, "reply", "r", false, "wait for a reply containing the results of the function execution")
	publishCmd.Flags().IntVarP(&publishOptions.count, "count", "c", 1, "the number of times to post the data")
	publishCmd.Flags().IntVarP(&publishOptions.pause, "pause", "p", 0, "the number of seconds to wait between postings")

	publishCmd.Flags().StringP("namespace", "", "default", "the namespace of the http-gateway")

	publishCmd.MarkFlagRequired("data")

}
