
package kafka

import (
	"github.com/Shopify/sarama"
	"log"
	"github.com/projectriff/message-transport/pkg/message"
)

func NewProducer(brokerAddrs []string) (*producer, error) {
	asyncProducer, err := sarama.NewAsyncProducer(brokerAddrs, nil)
	if err != nil {
		return &producer{}, err
	}

	errors := make(chan error)
	go func(errChan <-chan *sarama.ProducerError) {
		for {
			errors <- <-errChan
		}
	}(asyncProducer.Errors())

	return &producer{
		asyncProducer: asyncProducer,
		errors: errors,
	}, nil
}

type producer struct {
	asyncProducer sarama.AsyncProducer
	errors        chan error
}

func (p *producer) Send(topic string, message message.Message) error {
	kafkaMsg, err := toKafka(message)
	if err != nil {
		return err
	}
	kafkaMsg.Topic = topic

	p.asyncProducer.Input() <- kafkaMsg

	return nil
}


func (p *producer) Errors() <-chan error {
	return p.errors
}

func (p *producer) Close() error {
	err := p.asyncProducer.Close()
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

