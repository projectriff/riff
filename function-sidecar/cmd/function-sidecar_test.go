package main_test

import (
	"testing"
	"net/http"
	"os/exec"
	"os"
	"fmt"
	"gopkg.in/Shopify/sarama.v1"
	"bufio"
	"errors"
	"github.com/bsm/sarama-cluster"
	"math/rand"
	"time"
)

const sourceMsg = `World`
const expectedReply = `Hello World`

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestIntegrationWithKafka(t *testing.T) {

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		t.Fatal("Required environment variable KAFKA_BROKER was not provided")
	}

	cmd := exec.Command("../function-sidecar")
	input := randString(10)
	output := randString(10)
	group := randString(10)

	configJson := fmt.Sprintf(`{
		"spring.cloud.stream.kafka.binder.brokers":"%s",
		"spring.cloud.stream.bindings.input.destination": "%s",
		"spring.cloud.stream.bindings.output.destination": "%s",
		"spring.cloud.stream.bindings.input.group": "%s",
		"spring.profiles.active": "http"
	}`, broker, input, output, group)

	cmd.Env = []string{"SPRING_APPLICATION_JSON=" + configJson}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startErr := cmd.Start()
	defer cmd.Process.Kill()

	if startErr != nil {
		t.Fatal(startErr)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bodyScanner := bufio.NewScanner(r.Body)
		if ! bodyScanner.Scan() {
			t.Fatal(errors.New("Scan of message body failed"))
		}
		w.Write([]byte("Hello " + bodyScanner.Text()))
	})

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	kafkaProducer, kafkaProducerErr := sarama.NewAsyncProducer([]string{broker}, nil)
	if kafkaProducerErr != nil {
		t.Fatal(kafkaProducerErr)
	}

	testMessage := &sarama.ProducerMessage{Topic: input, Value: sarama.StringEncoder(string([]byte{0xff, 0x00}) + sourceMsg)}
	kafkaProducer.Input() <- testMessage
	producerCloseErr := kafkaProducer.Close()
	if producerCloseErr != nil {
		t.Fatal(producerCloseErr)
	}

	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	group2 := randString(10)
	consumer, err := cluster.NewConsumer([]string{broker}, group2, []string{output}, consumerConfig)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	select {
	case msg, ok := <-consumer.Messages():
		if ok {
			reply := string(msg.Value[2:])
			if reply != expectedReply {
				t.Fatal(fmt.Errorf("Received reply [%s] does not match expected reply [%s]", reply, expectedReply))
			}
		}
	case <-time.After(time.Second * 100):
		t.Fatal("Timed out waiting for reply")
	}
}
