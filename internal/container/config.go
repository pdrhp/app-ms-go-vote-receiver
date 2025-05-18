package container

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadConfig() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente existentes")
	}

	config := &Config{}

	config.Server.Port = getEnv("SERVER_PORT", "8080")

	config.Kafka.Brokers = strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	config.Kafka.Topic = getEnv("KAFKA_TOPIC", "votos")

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}