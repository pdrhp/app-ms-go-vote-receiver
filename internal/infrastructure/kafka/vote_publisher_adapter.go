package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/entity"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/port"
	"github.com/segmentio/kafka-go"
)

type VotePublisherAdapter struct {
	writer *kafka.Writer
}

func NewVotePublisherAdapter(brokers []string, topic string) (port.VotePublisher, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	adapter := &VotePublisherAdapter{
		writer: writer,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := adapter.HealthCheck(ctx); err != nil {
        writer.Close()
        return nil, fmt.Errorf("health check Kafka falhou: %w", err)
    }

	return adapter, nil
}

func (a *VotePublisherAdapter) PublishVote(ctx context.Context, vote *entity.Vote) error {
	message := map[string]interface{}{
		"id":            vote.ID,
		"participanteId": vote.ParticipantID,
		"sessionId":     vote.SessionID,
		"timestamp":     vote.Timestamp.Format(time.RFC3339),
	}

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return a.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(vote.SessionID),
		Value: value,
	})
}

func (a *VotePublisherAdapter) HealthCheck(ctx context.Context) error {
    conn, err := kafka.DialLeader(ctx, "tcp", a.writer.Addr.String(), a.writer.Topic, 0)
    if err != nil {
        return fmt.Errorf("falha ao conectar ao Kafka: %w", err)
    }
    defer conn.Close()

    conn.SetReadDeadline(time.Now().Add(5 * time.Second))

    brokers, err := conn.Brokers()
    if err != nil {
        return fmt.Errorf("falha ao acessar brokers Kafka: %w", err)
    }

    if len(brokers) == 0 {
        return fmt.Errorf("nenhum broker Kafka disponível")
    }

    return nil
}


func (a *VotePublisherAdapter) Close() error {
	return a.writer.Close()
}