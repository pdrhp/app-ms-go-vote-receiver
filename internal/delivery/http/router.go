package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pdrhp/ms-voto-receiver-go/internal/constants"
	"github.com/pdrhp/ms-voto-receiver-go/internal/core/usecase"
	"github.com/pdrhp/ms-voto-receiver-go/internal/delivery/http/handlers"
)

func SetupRouter(receiveVoteUseCase *usecase.ReceiveVoteUseCase) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    })

	voteHandler := handlers.NewVoteHandler(receiveVoteUseCase)

	v1 := router.Group(constants.APIVersionV1)
	{
		v1.POST(constants.EndpointVote, voteHandler.Handle)
	}

	return router
}