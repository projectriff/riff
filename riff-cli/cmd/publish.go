// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/kubectl"
	"github.com/projectriff/riff-cli/ioutils"
	"github.com/projectriff/riff-cli/minikube"
	"github.com/projectriff/riff-cli/jsonpath"
	"fmt"
	"net/http"
	"path/filepath"
	"os"
)

type PublishOptions struct {
	input string
	data  string
	reply bool
	count int16
	pause int16
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

		if publishOptions.data == "" {
			ioutils.Error("Missing required flag 'data'.")
			cmd.Usage()
			return;
		}

		if publishOptions.input == "" {
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			input := filepath.Dir(cwd)
			publishOptions.input = input
		}

		cmdArgs := []string{"get", "svc", "-l", "component=http-gateway", "-o", "json"}
		output, err := kubectl.QueryForBytes(cmdArgs)

		if err != nil {
			ioutils.Errorf("Error querying http-gateway %v\n %v\n", err, output)
			return
		}

		parser := jsonpath.NewParser(output)

		portType := parser.Value(`$.items[0].spec.type+`)

		if portType == "" {
			ioutils.Error("unable to locate http-gateway")
			return
		}

		port := parser.Value(`$.items[0].spec.ports[*]?(@.name == "http").nodePort+`)

		if port == "" {
			ioutils.Error("unable to determine http-gateway port")
			return
		}

		var ipAddress string

		switch portType {
		case "NodePort":
			ipAddress, err = minikube.QueryIp()
			if err != nil || strings.Contains(ipAddress, "Error getting IP") {
				ipAddress = "127.0.0.1"
			}
		case "LoadBalancer":
			ipAddress := parser.Value(`$.items[0].status.loadBalancer.ingress[0].ip+`)
			if ipAddress == "" {
				ioutils.Error("unable to determine http-gateway ip address")
				return
			}

		default:
			ioutils.Errorf("Unkown port type %s", portType)
			return
		}

		publish(ipAddress, port)

	},
}

func publish(ipAddress string, port string) {
	fmt.Printf("%s:%s\n",ipAddress, port)

	resource := "messages"
	if publishOptions.reply {
		resource = "requests"
	}
	
	url := fmt.Sprintf("http://%s:%s/%s/%s",ipAddress, port, resource, publishOptions.input )

	fmt.Printf("Posting to %s\n",url)

	resp, err := http.Post(url, "text/plain", strings.NewReader(publishOptions.data))
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

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
	publishCmd.Flags().StringVarP(&publishOptions.input, "input", "i", "", "The name of the input topic (defaults to the name of the current directory)")
	publishCmd.Flags().StringVarP(&publishOptions.data, "data", "d", "", "The data to post to the http-gateway using the input topic")
	publishCmd.Flags().BoolVarP(&publishOptions.reply, "reply", "r", false, "Wait for a reply containing the results of the function execution")
	publishCmd.Flags().Int16VarP(&publishOptions.count, "count", "c", 1, "The number of times to post the data")
	publishCmd.Flags().Int16VarP(&publishOptions.pause, "pause", "p", 0, "The number of seconds to wait between postings")

}
