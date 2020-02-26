package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	devutil "github.com/projectriff/developer-utils/pkg"
	client "github.com/projectriff/stream-client-go"
	"github.com/spf13/cobra"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	payload       string
	payloadBase64 string
	contentType   string
	header        []string
	namespace     string
)

func main() {
	if err := publishCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var publishCmd = &cobra.Command{
	Use:     "publish <stream-name>",
	Short:   "publish events to the given stream",
	Long:    "",
	Example: "publish letters --content-type text/plain --payload my-value",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		stop := devutil.SetupSignalHandler()
		go func() {
			select {
			case <-stop:
				cancel()
			}
		}()

		k8sClient := devutil.NewK8sClient()
		secretName, err := k8sClient.GetNestedString(args[0], namespace, devutil.StreamGVR, "status", "binding", "secretRef", "name")
		if err != nil {
			fmt.Println("error while finding binding secret reference", err)
			os.Exit(1)
		}

		encodedTopic, err := k8sClient.GetNestedString(secretName, namespace, devutil.SecretGVR, "data", "topic")
		if err != nil {
			fmt.Println("error while determining gateway topic for stream", err)
			os.Exit(1)
		}

		topic, err := base64.StdEncoding.DecodeString(encodedTopic)
		if err != nil {
			fmt.Println("error decoding topic", err)
			os.Exit(1)
		}

		encodedGateway, err := k8sClient.GetNestedString(secretName, namespace, devutil.SecretGVR, "data", "gateway")
		if err != nil {
			fmt.Println("error while determining gateway address for stream", err)
			os.Exit(1)
		}

		gateway, err := base64.StdEncoding.DecodeString(encodedGateway)
		if err != nil {
			fmt.Println("error decoding gateway address", err)
			os.Exit(1)
		}

		acceptableContentType, err := k8sClient.GetNestedString(args[0], namespace, devutil.StreamGVR, "spec", "contentType")
		if err != nil {
			fmt.Println("error while determining acceptableContentType for stream", err)
			os.Exit(1)
		}

		m, err := getMapFromHeaders(header)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		sc, err := client.NewStreamClient(string(gateway), string(topic), acceptableContentType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resolvedPayload, err := resolvePayload(payload, payloadBase64)
		_, err = sc.Publish(ctx, resolvedPayload, nil, contentType, m)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func resolvePayload(payload string, payloadBase64 string) (io.Reader, error) {
	if payload != "" && payloadBase64 != "" {
		return nil, errors.New("the options --payload and --payload-base64 are mutually exclusive")
	}
	if payloadBase64 != "" {
		payloadBytes, err := base64.StdEncoding.DecodeString(payloadBase64)
		if err != nil {
			return nil, errors.New("the payload is not base64 encoded")
		}
		return bytes.NewReader(payloadBytes), nil
	}
	return strings.NewReader(payload), nil
}

func init() {
	publishCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace of the stream")
	publishCmd.Flags().StringVarP(&payload, "payload", "p", "", "the content/payload to publish to stream")
	publishCmd.Flags().StringVarP(&contentType, "content-type", "c", "", "mime type of content")
	publishCmd.Flags().StringArrayVarP(&header, "header", "", header, "headers for the payload")
	publishCmd.Flags().StringVarP(&payloadBase64, "payload-base64", "b", "", "base64 encoded payload")
	err := publishCmd.MarkFlagRequired("content-type")
	if err != nil {
		panic(err)
	}
}

func getMapFromHeaders(headers []string) (map[string]string, error) {
	returnVal := map[string]string{}
	for _, h := range headers {
		splitH := strings.Split(h, ":")
		if len(splitH) != 2 {
			return nil, errors.New(fmt.Sprintf("illegal header: %s, expected form: k1:v1", h))
		}
		returnVal[splitH[0]] = splitH[1]
	}
	return returnVal, nil
}
