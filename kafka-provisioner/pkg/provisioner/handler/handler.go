package handler

import (
	"encoding/json"
	"fmt"
	client "github.com/projectriff/kafka-provisioner/pkg/provisioner/kafka"
	"io"
	"net/http"
	"strings"
)

type TopicCreationRequestHandler struct {
	KafkaClient client.KafkaClient
	Gateway     string
	Writer      io.Writer
}

func (rh *TopicCreationRequestHandler) GetHandlerFunc() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path[1:], "/")
		if len(parts) != 2 {
			responseWriter.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(responseWriter, "URLs should be of the form /<namespace>/<stream-name>\n")
			return
		}
		// NOTE: choice of underscore as separator is important as it is not allowed in k8s names
		topicName := fmt.Sprintf("%s_%s", parts[0], parts[1])
		topicExists, kafkaError := rh.KafkaClient.TopicExists(topicName)
		if kafkaError != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			if err := kafkaError.GeneralError; err != nil {
				_, _ = fmt.Fprintf(rh.Writer, "Error trying to list topics to see if %q exists: %v\n", topicName, err)
				_, _ = fmt.Fprintf(responseWriter, "Error trying to list topics to see if %q exists: %v\n", topicName, err)
				return
			}

			kafkaErrorCode := kafkaError.KError
			_, _ = fmt.Fprintf(rh.Writer, "Error trying to list topics to see if %q exists: %v\n", topicName, kafkaErrorCode)
			_, _ = fmt.Fprintf(responseWriter, "Error trying to list topics to see if %q exists: %v\n", topicName, kafkaErrorCode)
			return
		}
		if !topicExists {
			if err := rh.KafkaClient.CreateTopic(topicName); err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(rh.Writer, "Error creating topic %q: %v\n", topicName, err)
				_, _ = fmt.Fprintf(responseWriter, "Error creating topic %q: %v\n", topicName, err)
				return
			}
			responseWriter.WriteHeader(http.StatusCreated)
		} else {
			responseWriter.WriteHeader(http.StatusOK)
		}

		if err := encodeResponse(responseWriter, rh.Gateway, topicName); err != nil {
			_, _ = fmt.Fprintf(rh.Writer, "Failed to write json response: %v", err)
			return
		}
		_, _ = fmt.Fprintf(rh.Writer, "Reported successful topic %q\n", topicName)
	}
}

func encodeResponse(w http.ResponseWriter, gateway string, topicName string) error {
	w.Header().Set("Content-Type", "application/json")
	res := result{
		Gateway: gateway,
		Topic:   topicName,
	}
	return json.NewEncoder(w).Encode(res)
}

type result struct {
	Gateway string `json:"gateway"`
	Topic   string `json:"topic"`
}
