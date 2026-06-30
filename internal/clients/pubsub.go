package clients

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPubSubPublisher(ctx context.Context, projectID, topicID string) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error creando cliente pubsub v2: %w", err)
	}

	topic := client.Topic(topicID)

	return &Publisher{
		client: client,
		topic:  topic,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, payload []byte) error {
	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: payload,
	})

	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("error al recibir confirmación (ACK) del bus v2: %w", err)
	}

	return nil
}

func (p *Publisher) Close() error {
	p.topic.Stop()
	return p.client.Close()
}
