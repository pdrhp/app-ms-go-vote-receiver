package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/usecase"
)

type VoteRequest struct {
    ParticipanteID int    `json:"participanteId" binding:"required"`
    SessionID      string `json:"sessionId" binding:"required"`
}

type VoteHandler struct {
	receiveVoteUseCase *usecase.ReceiveVoteUseCase
}

func NewVoteHandler(receiveVoteUseCase *usecase.ReceiveVoteUseCase) *VoteHandler {
	return &VoteHandler{
		receiveVoteUseCase: receiveVoteUseCase,
	}
}

func (h *VoteHandler) Handle(c *gin.Context) {
	start := time.Now()

	var request usecase.ReceiveVoteInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Payload inválido",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Recebida requisição de voto: participanteId=%d, sessionId=%s", request.ParticipantID, request.SessionID)

	output, err := h.receiveVoteUseCase.Execute(c.Request.Context(), request)

	duration := time.Since(start)

	if err != nil {
		log.Printf("Erro após %v: %v", duration, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Falha ao processar voto",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Voto recebido com sucesso",
		"voteId":    output.VoteID,
		"timestamp": output.Timestamp,
	})
}