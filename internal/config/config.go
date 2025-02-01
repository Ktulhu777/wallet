package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env string `env:"ENV" env-default:"local" env-required:"true"`
	Storage
	HTTPServer
}

type HTTPServer struct {
	Address     string        `env:"SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"SERVER_TIMEOUT" env-required:"true"`
	IdleTimeout time.Duration `env:"SERVER_IDLE_TIMEOUT" env-required:"true"`
}

type Storage struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     int    `env:"DB_PORT" env-default:"5432"`
	DBName   string `env:"DB_NAME" env-required:"true"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASS" env-required:"true"`
	SSLMode  string `env:"DB_SSL_MODE" env-default:"disable"`
}

func MustLoad() *Config {
	var envFilePath string = "./config/config.env"
	if err := godotenv.Load(envFilePath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
