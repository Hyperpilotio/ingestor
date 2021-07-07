package database

import (
	"log"

	kafka "github.com/Shopify/sarama"
	// "github.com/hyperpilotio/ingestor/config"
	// "github.com/hyperpilotio/ingestor/log"
)

var producer kafka.SyncProducer

func init() {
	// FIXME pass configuration into this function
	producer = newKafkaProducer()
}

// func newKafkaProducer(config config.Provider) {}
func newKafkaProducer() kafka.SyncProducer {
	// FIXME Change the fixed value
	producer, err := kafka.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalln(err)
	}
	return producer
}

func Producer() *kafka.SyncProducer {
	return &producer
}

func CloseProducer() {
	if err := producer.Close(); err != nil {
		log.Fatalln(err)
	}
}

func Send(topic, value string) {
	msg := &kafka.ProducerMessage{Topic: topic, Value: kafka.StringEncoder(value)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}
}
