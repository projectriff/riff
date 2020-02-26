package handler_test

import (
	"fmt"
	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/kafka-provisioner/pkg/provisioner/handler"
	client "github.com/projectriff/kafka-provisioner/pkg/provisioner/kafka"
	"github.com/projectriff/kafka-provisioner/pkg/provisioner/kafka/kafkafakes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Provisioner HTTP Handler", func() {

	const (
		gateway                = "liiklus.example.com"
		existingTopicNamespace = "some-namespace"
		existingTopicName      = "some-topic"
	)

	var (
		kafkaTopicName      = fmt.Sprintf("%s_%s", existingTopicNamespace, existingTopicName)
		responseRecorder    *httptest.ResponseRecorder
		fakeKafkaClient     *kafkafakes.FakeKafkaClient
		creationHandlerFunc http.HandlerFunc
		request             *http.Request
	)

	BeforeEach(func() {
		responseRecorder = httptest.NewRecorder()
		fakeKafkaClient = &kafkafakes.FakeKafkaClient{}
		request = putRequest(fmt.Sprintf("/%s/%s", existingTopicNamespace, existingTopicName))
		creationHandler := &handler.TopicCreationRequestHandler{
			KafkaClient: fakeKafkaClient,
			Gateway:     gateway,
			Writer:      ioutil.Discard}
		creationHandlerFunc = creationHandler.GetHandlerFunc()
	})

	It("returns 200 if the topic already exists", func() {
		fakeKafkaClient.TopicExistsReturns(true, nil)

		creationHandlerFunc.ServeHTTP(responseRecorder, request)

		Expect(responseRecorder.Code).To(Equal(http.StatusOK),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusOK, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).To(MatchJSON(
			fmt.Sprintf("{"+
				"	\"gateway\":\"%s\","+
				"	\"topic\":\"%s_%s\""+
				"}", gateway, existingTopicNamespace, existingTopicName)))
	})

	It("returns 201 if the topic is successfully created", func() {
		fakeKafkaClient.TopicExistsReturns(false, nil)
		fakeKafkaClient.CreateTopicReturns(nil)

		creationHandlerFunc.ServeHTTP(responseRecorder, request)

		Expect(responseRecorder.Code).To(Equal(http.StatusCreated),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusCreated, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).To(MatchJSON(
			fmt.Sprintf(`{"gateway": "%s", "topic": "%s_%s"}`, gateway, existingTopicNamespace, existingTopicName)))
	})

	It("returns 400 if the the topic is not properly specified", func() {
		creationHandlerFunc.ServeHTTP(responseRecorder, putRequest("/invalid-topic"))

		Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusBadRequest, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).
			To(Equal("URLs should be of the form /<namespace>/<stream-name>\n"))
	})

	It("returns 500 if an unexpected error occurred while listing topics", func() {
		fakeKafkaClient.TopicExistsReturns(false, &client.KafkaError{GeneralError: fmt.Errorf("oopsie")})

		creationHandlerFunc.ServeHTTP(responseRecorder, request)

		Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusInternalServerError, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).
			To(Equal("Error trying to list topics to see if \"" + kafkaTopicName + "\" exists: oopsie\n"))
	})

	It("returns 500 if a server error occurred while listing topics", func() {
		fakeKafkaClient.TopicExistsReturns(false, &client.KafkaError{KError: sarama.ErrInvalidPartitions})

		creationHandlerFunc.ServeHTTP(responseRecorder, request)

		Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusInternalServerError, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).
			To(Equal("Error trying to list topics to see if \"" + kafkaTopicName + "\" exists: kafka server: Number of partitions is invalid.\n"))
	})

	It("returns 500 if an error occurred while creating a topic", func() {
		fakeKafkaClient.TopicExistsReturns(false, nil)
		fakeKafkaClient.CreateTopicReturns(fmt.Errorf("oopsie"))

		creationHandlerFunc.ServeHTTP(responseRecorder, request)

		Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError),
			fmt.Sprintf("Expected %d after topic creation request but got %d", http.StatusInternalServerError, responseRecorder.Code))
		Expect(responseRecorder.Body.String()).
			To(Equal("Error creating topic \"" + kafkaTopicName + "\": oopsie\n"))
	})
})

func putRequest(path string) *http.Request {
	return httptest.NewRequest("PUT", path, nil)
}
