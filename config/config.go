package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var Cfg *Config

type Config struct {
	ServerPort    int    `env:"SERVER_PORT" envDefault:"8080"`
	RegistryType  string `env:"REGISTRY_TYPE" envDefault:"memory"` // memory or redis
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`
	DefaultTTL    int    `env:"DEFAULT_TTL" envDefault:"0"`

	DefaultFuncConcurrency int `env:"DEFAULT_FUNC_CONCURRENCY" envDefault:"4"`
	MaxNodeConcurrency     int `env:"MAX_NODE_CONCURRENCY" envDefault:"0"`
	QueueSizePerFunction   int `env:"QUEUE_SIZE_PER_FUNCTION" envDefault:"128"`
	InvokeTimeoutMillis    int `env:"INVOKE_TIMEOUT_MS" envDefault:"5000"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system env vars")
	}
	Cfg = &Config{}
	if err := env.Parse(Cfg); err != nil {
		return nil, err
	}
	return Cfg, nil
}
