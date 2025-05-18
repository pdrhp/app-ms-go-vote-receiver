package port

import (
	"context"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/entity"
)

type VotePublisher interface {
	PublishVote(ctx context.Context, vote *entity.Vote) error
}
