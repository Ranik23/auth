package mock_reader

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReaderInterface interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type MockKafkaReaderImpl struct{}

func (m *MockKafkaReaderImpl) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return kafka.Message{}, nil
}

func (m *MockKafkaReaderImpl) Close() error {
	return nil
}