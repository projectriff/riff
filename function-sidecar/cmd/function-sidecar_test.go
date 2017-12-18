package main_test

import (
	"testing"
	"net/http"
	"os/exec"
	"os"
	"fmt"
	"bufio"
	"errors"
	"github.com/bsm/sarama-cluster"
	"math/rand"
	"time"
	"github.com/projectriff/function-sidecar/pkg/wireformat"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"github.com/Shopify/sarama"
)

const sourceMsg = `World`
const expectedReply = `Hello World`

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	input := randString(10)
	output := randString(10)
	group := randString(10)
	cmd := exec.Command("../function-sidecar", "--inputs", input, "--outputs", output, "--brokers", broker, "--group", group, "--protocol", "http")

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

	testMessage, err := wireformat.ToKafka(dispatcher.Message{Payload: []byte(sourceMsg)})
	if err != nil {
		t.Fatal(err)
	}
	testMessage.Topic = input
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
	case consumerMessage, ok := <-consumer.Messages():
		if ok {
			msg, err := wireformat.FromKafka(consumerMessage)
			if err != nil {
				t.Fatal(err)
			}
			reply := string(msg.Payload)
			if reply != expectedReply {
				t.Fatal(fmt.Errorf("Received reply [%s] does not match expected reply [%s]", reply, expectedReply))
			}
		}
	case <-time.After(time.Second * 100):
		t.Fatal("Timed out waiting for reply")
	}
}
