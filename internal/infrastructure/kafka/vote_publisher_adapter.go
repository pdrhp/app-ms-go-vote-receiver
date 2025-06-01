package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/entity"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/port"
	"github.com/segmentio/kafka-go"
)

type VotePublisherAdapter struct {
	writers    []*kafka.Writer
	brokers    []string
	topic      string
	poolSize   int
	current    int
	mutex      sync.RWMutex
}

func NewVotePublisherAdapter(brokers []string, topic string) (port.VotePublisher, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("pelo menos um broker deve ser configurado")
	}

	poolSize := 3
	writers := make([]*kafka.Writer, poolSize)

	for i := 0; i < poolSize; i++ {
		writers[i] = &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,

			Balancer:     &kafka.RoundRobin{},

			BatchSize:    1000,
			BatchBytes: 	16_000_000,
			BatchTimeout: 30 * time.Millisecond,

			RequiredAcks: kafka.RequireOne,

			Async:        true,

			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,

			MaxAttempts:  2,

			Compression:  kafka.Lz4,

			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					log.Printf("Erro ao produzir mensagens: %v", err)
					return
				}

				log.Printf("Mensagens produzidas com sucesso: %d", len(messages))
			},
		}
	}

	adapter := &VotePublisherAdapter{
		writers:  writers,
		brokers:  brokers,
		topic:    topic,
		poolSize: poolSize,
		current:  0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Iniciando health check para %d broker(s): %v", len(brokers), brokers)
	if err := adapter.HealthCheck(ctx); err != nil {
		for _, writer := range writers {
			writer.Close()
		}
		return nil, fmt.Errorf("health check Kafka falhou: %w", err)
	}

	log.Printf("Health check concluído com sucesso para todos os brokers")
	return adapter, nil
}

func (a *VotePublisherAdapter) getWriter() *kafka.Writer {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	writer := a.writers[a.current]
	a.current = (a.current + 1) % a.poolSize
	return writer
}

func (a *VotePublisherAdapter) PublishVote(ctx context.Context, vote *entity.Vote) error {
	writerIndex := a.current
	writer := a.getWriter()

	message := map[string]interface{}{
		"id":            vote.ID,
		"participanteId": vote.ParticipantID,
		"sessionId":     vote.SessionID,
		"timestamp":     vote.Timestamp.Format(time.RFC3339),
	}

	value, err := json.Marshal(message)
	if err != nil {
		log.Printf("Erro ao serializar mensagem: %v", err)
		return err
	}

	log.Printf("[Pool-%d] Enviando mensagem: %s", writerIndex, vote.ID[:8])

	kafkaMessage := kafka.Message{
		Key:   []byte(vote.SessionID),
		Value: value,
	}

	log.Printf("Enviando para brokers: %v, tópico: %s (writer pool)", a.brokers, a.topic)

	if err := writer.WriteMessages(ctx, kafkaMessage); err != nil {
		log.Printf("Erro ao publicar mensagem no Kafka: %v", err)
		return err
	}

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
			log.Printf("Falha ao conectar com broker %s: %v", broker, err)
			lastErr = err
			cancel()
			continue
		}

		_, err = conn.ReadPartitions(a.topic)
		conn.Close()
		cancel()

		if err != nil {
			log.Printf("Falha ao acessar tópico '%s' no broker %s: %v", a.topic, broker, err)
			lastErr = err
			continue
		}

		log.Printf("Broker %s está funcionando corretamente", broker)
		successfulBrokers++
	}

	if successfulBrokers == 0 {
		return fmt.Errorf("nenhum broker está acessível. Último erro: %w", lastErr)
	}

	if successfulBrokers < len(a.brokers) {
		log.Printf("Aviso: %d de %d brokers estão funcionando", successfulBrokers, len(a.brokers))
	}

	log.Printf("Health check concluído: %d/%d brokers funcionando", successfulBrokers, len(a.brokers))
	return nil
}


func (a *VotePublisherAdapter) Close() error {
	log.Printf("Fechando pool de %d writers", a.poolSize)

	var lastErr error
	for i, writer := range a.writers {
		if err := writer.Close(); err != nil {
			log.Printf("Erro ao fechar writer %d: %v", i, err)
			lastErr = err
		}
	}

	return lastErr
}