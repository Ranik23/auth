package mock_writer

import (
	"context"

	"github.com/segmentio/kafka-go"
)



type KafkaWriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error 
}


type MockKafkaWriterImpl struct {}


func (m *MockKafkaWriterImpl) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return nil
}


func (m *MockKafkaWriterImpl) Close() error {
	return nil
}