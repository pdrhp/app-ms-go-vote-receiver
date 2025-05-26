package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/entity"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/port"
	"github.com/segmentio/kafka-go"
)

type VotePublisherAdapter struct {
	writer  *kafka.Writer
	brokers []string
	topic   string
}

func NewVotePublisherAdapter(brokers []string, topic string) (port.VotePublisher, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("pelo menos um broker deve ser configurado")
	}

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
		writer:  writer,
		brokers: brokers,
		topic:   topic,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Iniciando health check para %d broker(s): %v", len(brokers), brokers)
	if err := adapter.HealthCheck(ctx); err != nil {
		writer.Close()
		return nil, fmt.Errorf("health check Kafka falhou: %w", err)
	}

	log.Printf("Health check concluído com sucesso para todos os brokers")
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

	log.Printf("Enviando mensagem para Kafka: %v", message)

	if err := a.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(vote.SessionID),
		Value: value,
	}); err != nil {
		log.Printf("Erro ao publicar mensagem no Kafka: %v", err)
		return err
	}

	log.Println("Mensagem publicada com sucesso no Kafka")
	return nil
}

func (a *VotePublisherAdapter) HealthCheck(ctx context.Context) error {
	var lastErr error
	successfulBrokers := 0

  for i, broker := range a.brokers {
		log.Printf("Testando conexão com broker %d/%d: %s", i+1, len(a.brokers), broker)

		brokerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

		conn, err := kafka.DialContext(brokerCtx, "tcp", broker)
		if err != nil {
			log.Printf("❌ Falha ao conectar com broker %s: %v", broker, err)
			lastErr = err
			cancel()
			continue
		}

		_, err = conn.ReadPartitions(a.topic)
		conn.Close()
		cancel()

		if err != nil {
			log.Printf("❌ Falha ao acessar tópico '%s' no broker %s: %v", a.topic, broker, err)
			lastErr = err
			continue
		}

		log.Printf("✅ Broker %s está funcionando corretamente", broker)
		successfulBrokers++
	}

	if successfulBrokers == 0 {
		return fmt.Errorf("nenhum broker está acessível. Último erro: %w", lastErr)
	}

	if successfulBrokers < len(a.brokers) {
		log.Printf("⚠️  Aviso: %d de %d brokers estão funcionando", successfulBrokers, len(a.brokers))
	}

	log.Printf("Health check concluído: %d/%d brokers funcionando", successfulBrokers, len(a.brokers))
	return nil
}


func (a *VotePublisherAdapter) Close() error {
	return a.writer.Close()
}