package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/pdrhp/ms-voto-receiver-go/internal/container"
	"github.com/pdrhp/ms-voto-receiver-go/internal/delivery/http"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
  debug.SetGCPercent(100)

	log.Printf("Configurado para usar %d CPUs", runtime.NumCPU())

	config := container.LoadConfig()

	c, err := container.NewContainer(config)
	if err != nil {
		log.Fatalf("Erro ao criar container: %v", err)
	}

	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("Erro ao fechar container: %v", err)
		}
	}()

	router := http.SetupRouter(c.ReceiveVoteUseCase)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverAddr := fmt.Sprintf(":%s", config.Server.Port)
	go func() {
		log.Printf("Servidor iniciado na porta %s", config.Server.Port)
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	<-quit
	log.Println("Servidor encerrado")
}