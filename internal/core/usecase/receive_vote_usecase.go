package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/entity"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/port"
)

type ReceiveVoteUseCase struct {
	voteRepo port.VotePublisher
}

func NewReceiveVoteUseCase(voteRepo port.VotePublisher) *ReceiveVoteUseCase {
	return &ReceiveVoteUseCase{
		voteRepo: voteRepo,
	}
}

type ReceiveVoteInput struct {
	ParticipantID int    `json:"participanteId" binding:"required"`
	SessionID     string `json:"sessionId" binding:"required"`
}

type ReceiveVoteOutput struct {
	VoteID    string    `json:"voteId"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

func (uc *ReceiveVoteUseCase) Execute(ctx context.Context, input ReceiveVoteInput) (*ReceiveVoteOutput, error) {
	if input.ParticipantID <= 0 {
		return nil, fmt.Errorf("ID do participante deve ser um número positivo")
	}

	if input.SessionID == "" {
		return nil, fmt.Errorf("session ID não pode ser vazio")
	}

	log.Printf("Criando voto: participanteId=%d, sessionId=%s", input.ParticipantID, input.SessionID)

	vote := entity.NewVote(input.ParticipantID, input.SessionID)

	err := uc.voteRepo.PublishVote(context.Background(), vote)

	if err != nil {
		vote.MarkAsFailed()
		return nil, fmt.Errorf("falha ao publicar voto: %w", err)
	}

	vote.MarkAsSent()

	return &ReceiveVoteOutput{
		VoteID:    vote.ID,
		Timestamp: vote.Timestamp,
		Status:    string(vote.Status),
	}, nil
}