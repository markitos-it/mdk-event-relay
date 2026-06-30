package clients

import (
	"context"
	"log"
)

type ConsolePublisher struct{}

func NewConsolePublisher() *ConsolePublisher {
	return &ConsolePublisher{}
}

func (c *ConsolePublisher) Publish(ctx context.Context, payload []byte) error {
	log.Printf("[MOCK-BUS] Evento capturado: %s", string(payload))

	return nil
}

func (c *ConsolePublisher) Close() error {
	return nil
}
