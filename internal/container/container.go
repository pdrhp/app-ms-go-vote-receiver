package container

import (
	"log"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/port"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/usecase"
	"github.com/pdrhp/ms-voto-receiver-go/internal/infrastructure/kafka"
)

type Container struct {
	VoteRepository port.VotePublisher

	ReceiveVoteUseCase *usecase.ReceiveVoteUseCase
}

func NewContainer(config *Config) (*Container, error) {
	c := &Container{}

	log.Println("Inicializando adaptador Kafka...")

	publisher, err := kafka.NewVotePublisherAdapter(
		config.Kafka.Brokers,
		config.Kafka.Topic,
	)
	if err != nil {
		return nil, err
	}

	c.VoteRepository = publisher

	log.Println("Inicializando caso de uso ReceiveVoteUseCase...")
	c.ReceiveVoteUseCase = usecase.NewReceiveVoteUseCase(c.VoteRepository)

	return c, nil
}

func (c *Container) Close() error {
	log.Println("Fechando recursos do container...")

	if closer, ok := c.VoteRepository.(interface{ Close() error }); ok {
		return closer.Close()
	}

	return nil
}

type Config struct {
	Server struct {
		Port string
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
}