package clients

import (
	"context"
	"fmt"

	"markitos-it-event-relay/internal/domain"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
)

type PubSubPublisher struct {
	client  *pubsub.Client
	topicID string
}

func NewPubSubPublisher(ctx context.Context, projectID, topicID string, opts ...option.ClientOption) (domain.EventPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating pubsub client: %w", err)
	}

	return &PubSubPublisher{
		client:  client,
		topicID: topicID,
	}, nil
}

func (p *PubSubPublisher) Publish(ctx context.Context, payload []byte) error {
	publisher := p.client.Publisher(p.topicID)

	msg := &pubsub.Message{
		Data: payload,
	}

	result := publisher.Publish(ctx, msg)

	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("error getting result from pubsub: %w", err)
	}

	return nil
}

func (p *PubSubPublisher) Close() error {
	return p.client.Close()
}
