package http

import (
	"github.com/gin-gonic/gin"
	"github.com/pdrhp/ms-voto-receiver-go/internal/constants"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/usecase"
	"github.com/pdrhp/ms-voto-receiver-go/internal/delivery/http/handlers"
)

func SetupRouter(receiveVoteUseCase *usecase.ReceiveVoteUseCase) *gin.Engine {
	router := gin.Default()

	voteHandler := handlers.NewVoteHandler(receiveVoteUseCase)

	v1 := router.Group(constants.APIVersionV1)
	{
		v1.POST(constants.EndpointVote, voteHandler.Handle)
	}

	return router
}