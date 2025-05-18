package container

import (
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

	publisher, err := kafka.NewVotePublisherAdapter(
		config.Kafka.Brokers,
		config.Kafka.Topic,
	)
	if err != nil {
		return nil, err
	}
	c.VoteRepository = publisher

	c.ReceiveVoteUseCase = usecase.NewReceiveVoteUseCase(c.VoteRepository)

	return c, nil
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