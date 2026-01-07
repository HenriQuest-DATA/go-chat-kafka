package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Worker   WorkerConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type KafkaConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	RetryMax      int
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type WorkerConfig struct {
	PoolSize       int
	BufferSize     int
	ProcessTimeout time.Duration
}

// Load carrega as configurações do .env
func Load() (*Config, error) {
	_ = godotenv.Load()

	// Validar TODAS as variáveis obrigatórias de uma vez
	requiredEnvVars := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"KAFKA_BROKERS",
		"KAFKA_TOPIC",
		"KAFKA_CONSUMER_GROUP",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			return nil, fmt.Errorf("variável de ambiente obrigatória não definida: %s", envVar)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     parseDuration(getEnv("SERVER_READ_TIMEOUT", "15s")),
			WriteTimeout:    parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "15s")),
			ShutdownTimeout: parseDuration(getEnv("SHUTDOWN_TIMEOUT", "30s")),
		},
		Database: DatabaseConfig{
			Host:            os.Getenv("DB_HOST"),
			Port:            os.Getenv("DB_PORT"),
			User:            os.Getenv("DB_USER"),
			Password:        os.Getenv("DB_PASSWORD"),
			DBName:          os.Getenv("DB_NAME"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25")),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "5")),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m")),
		},
		Kafka: KafkaConfig{
			Brokers:       strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
			Topic:         os.Getenv("KAFKA_TOPIC"),
			ConsumerGroup: os.Getenv("KAFKA_CONSUMER_GROUP"),
			RetryMax:      parseInt(getEnv("KAFKA_RETRY_MAX", "3")),
		},
		JWT: JWTConfig{
			AccessSecret:      os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret:     os.Getenv("JWT_REFRESH_SECRET"),
			AccessExpiration:  1 * time.Hour,
			RefreshExpiration: 7 * 24 * time.Hour,
		},
		Worker: WorkerConfig{
			PoolSize:       parseInt(getEnv("WORKER_POOL_SIZE", "10")),
			BufferSize:     parseInt(getEnv("WORKER_BUFFER_SIZE", "100")),
			ProcessTimeout: parseDuration(getEnv("WORKER_TIMEOUT", "30s")),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifica configurações obrigatórias
func (c *Config) Validate() error {
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET é obrigatório")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET é obrigatório")
	}
	return nil
}

// DSN retorna string de conexão PostgreSQL
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Funções auxiliares
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
