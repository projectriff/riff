package client_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	client "github.com/projectriff/riff/stream-client-go"
	"github.com/projectriff/riff/stream-client-go/pkg/liiklus"
)

// This is an integration test meant to be run against a liiklus gateway. Please refer to the CI scripts for
// setup details
func TestSimplePublishSubscribe(t *testing.T) {
	now := time.Now()
	topic := topicName(t.Name(), fmt.Sprintf("%d%d%d", now.Hour(), now.Minute(), now.Second()))

	c := setupStreamingClient(topic, t)

	payload := "FOO"
	headers := map[string]string{"H1": "V1", "H2": "V2"}
	publish(c, payload, "text/plain", topic, headers, t)
	subscribe(c, payload, topic, true, headers, t)
}

func setupStreamingClient(topic string, t *testing.T) *client.StreamClient {
	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}
	return c
}

func publish(c *client.StreamClient, value, contentType, topic string, headers map[string]string, t *testing.T) {
	reader := strings.NewReader(value)
	publishResult, err := c.Publish(context.Background(), reader, nil, contentType, headers)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Published: %+v\n", publishResult)
}

func subscribe(c *client.StreamClient, expectedValue, topic string, fromBeginning bool, headers map[string]string, t *testing.T) {

	var errHandler client.EventErrHandler
	errHandler = func(cancel context.CancelFunc, err error) {
		fmt.Printf("cancelling subscriber due to: %v", err)
		cancel()
	}

	payloadChan := make(chan string)
	headersChan := make(chan map[string]string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, event liiklus.LiiklusEvent) error {
		payloadChan <- string(event.Data)
		headersChan <- headers
		if event.Id == "" {
			t.Error("did not expect id to be empty")
		}
		if event.Time == "" {
			t.Error("did not expect time to be empty")
		}
		return nil
	}

	_, err := c.Subscribe(context.Background(), "g8", fromBeginning, eventHandler, errHandler)
	if err != nil {
		t.Error(err)
	}
	v1 := <-payloadChan
	if v1 != expectedValue {
		t.Errorf("expected value: %s, but was: %s", expectedValue, v1)
	}
	// see: https://github.com/projectriff/stream-client-go/issues/19
	//h := <-headersChan
	//if !reflect.DeepEqual(headers, h) {
	//	t.Errorf("headers not equal. expected %s, but was: %s", headers, h)
	//}
}

func TestSubscribeBeforePublish(t *testing.T) {
	now := time.Now()
	topic := topicName(t.Name(), fmt.Sprintf("%d%d%d", now.Hour(), now.Minute(), now.Second()))

	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}

	testVal := "BAR"
	result := make(chan string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, event liiklus.LiiklusEvent) error {
		result <- string(event.Data)
		return nil
	}
	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		t.Error("Did not expect an error")
	}
	_, err = c.Subscribe(context.Background(), t.Name(), true, eventHandler, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	publish(c, testVal, "text/plain", topic, nil, t)
	v1 := <-result
	if v1 != testVal {
		t.Errorf("expected value: %s, but was: %s", testVal, v1)
	}
}

func TestSubscribeCancel(t *testing.T) {
	now := time.Now()
	topic := topicName(t.Name(), fmt.Sprintf("%d%d%d", now.Hour(), now.Minute(), now.Second()))

	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}

	expectedError := "expected_error"
	result := make(chan string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, event liiklus.LiiklusEvent) error {
		result <- string(event.Data)
		return nil
	}
	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		result <- expectedError
	}
	cancel, err := c.Subscribe(context.Background(), t.Name(), true, eventHandler, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	cancel()
	v1 := <-result
	if v1 != expectedError {
		t.Errorf("expected value: %s, but was: %s", expectedError, v1)
	}
}

func TestMultipleSubscribe(t *testing.T) {
	now := time.Now()
	topic1 := topicName(t.Name(), fmt.Sprintf("%d%d%d_1", now.Hour(), now.Minute(), now.Second()))
	topic2 := topicName(t.Name(), fmt.Sprintf("%d%d%d_2", now.Hour(), now.Minute(), now.Second()))

	c1 := setupStreamingClient(topic1, t)
	c2 := setupStreamingClient(topic2, t)

	testVal1 := "BAR1"
	testVal2 := "BAR2"
	result1 := make(chan string)
	result2 := make(chan string)

	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		panic(err)
	}
	var err error
	_, err = c1.Subscribe(context.Background(), t.Name()+"1", true, func(ctx context.Context, event liiklus.LiiklusEvent) error {
		result1 <- string(event.Data)
		return nil
	}, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	_, err = c2.Subscribe(context.Background(), t.Name()+"2", true, func(ctx context.Context, event liiklus.LiiklusEvent) error {
		result2 <- string(event.Data)
		return nil
	}, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	publish(c1, testVal1, "text/plain", topic1, nil, t)
	publish(c2, testVal2, "text/plain", topic1, nil, t)

	v1 := <-result1
	if v1 != testVal1 {
		t.Errorf("expected value: %s, but was: %s", testVal1, v1)
	}
	v2 := <-result2
	if v2 != testVal2 {
		t.Errorf("expected value: %s, but was: %s", testVal2, v2)
	}
}

func TestSubscribeFromLatest(t *testing.T) {
	now := time.Now()
	topic := topicName(t.Name(), fmt.Sprintf("%d%d%d", now.Hour(), now.Minute(), now.Second()))

	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}
	testVal1 := "testVal1"
	testVal2 := "testVal2"
	result := make(chan string, 1)

	publish(c, testVal1, "text/plain", topic, nil, t)

	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		panic(err)
	}
	_, err = c.Subscribe(context.Background(), t.Name(), false, func(ctx context.Context, event liiklus.LiiklusEvent) error {
		result <- string(event.Data)
		return nil
	}, eventErrHandler)
	if err != nil {
		t.Fatal(err)
	}
	// subscribe goroutine may not have entered Recv() before the event is published
	time.Sleep(5 * time.Second)

	publish(c, testVal2, "text/plain", topic, nil, t)
	v := <-result
	if v != testVal2 {
		t.Errorf("expected value: %s, but was: %s", testVal2, v)
	}
}

func topicName(namespace, name string) string {
	switch os.Getenv("GATEWAY") {
	case "pulsar":
		return fmt.Sprintf("persistent://public/default/%s-%s", namespace, name)
	default:
		return fmt.Sprintf("%s_%s", namespace, name)
	}
}
