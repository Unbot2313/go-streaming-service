package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type MockRabbitMQService struct {
	ConnectFn func() error
	CloseFn   func()
	PublishFn func(queueName string, message []byte) error
	ConsumeFn func(queueName string, handler services.MessageHandler) error
}

func (m *MockRabbitMQService) Connect() error {
	return m.ConnectFn()
}

func (m *MockRabbitMQService) Close() {
	m.CloseFn()
}

func (m *MockRabbitMQService) Publish(queueName string, message []byte) error {
	return m.PublishFn(queueName, message)
}

func (m *MockRabbitMQService) Consume(queueName string, handler services.MessageHandler) error {
	return m.ConsumeFn(queueName, handler)
}
