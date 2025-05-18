package entity

import (
	"fmt"
	"time"

	"github.com/pdrhp/ms-voto-receiver-go/internal/core/util"
)

type Vote struct {
	ID            string
	ParticipantID int
	SessionID     string
	Timestamp     time.Time
	Status        VoteStatus
}

func NewVote(participantID int, sessionID string) *Vote {
	return &Vote{
		ID:            util.GenerateUUID(),
		ParticipantID: participantID,
		SessionID:     sessionID,
		Timestamp:     time.Now().UTC(),
		Status:        VoteStatusReceived,
	}
}

func (v *Vote) SetStatus(status VoteStatus) error {
	if err := status.Validate(); err != nil {
		return fmt.Errorf("falha ao alterar status: %w", err)
	}
	v.Status = status
	return nil
}

func (v *Vote) MarkAsSent() {
	v.Status = VoteStatusSent
}


func (v *Vote) MarkAsFailed() {
	v.Status = VoteStatusFailed
}

func (v *Vote) IsProcessed() bool {
	return v.Status == VoteStatusSent || v.Status == VoteStatusFailed
}

func (v *Vote) String() string {
	return fmt.Sprintf("Vote[ID=%s, ParticipantID=%d, Status=%s]",
		v.ID, v.ParticipantID, v.Status)
}