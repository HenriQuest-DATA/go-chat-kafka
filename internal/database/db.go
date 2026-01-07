package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"chat-kafka-go/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// New cria nova conexão com PostgreSQL
func New(ctx context.Context, cfg *config.DatabaseConfig) (*DB, error) {
	// Parse config
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear config: %w", err)
	}

	// Configurar pool de conexões
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Conectar
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar: %w", err)
	}

	// Testar conexão
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("falha no ping: %w", err)
	}

	log.Println("✓ Database conectado com sucesso")
	return &DB{Pool: pool}, nil
}

// Close fecha conexão
func (db *DB) Close() {
	db.Pool.Close()
	log.Println("✓ Database desconectado")
}

// Health verifica saúde do banco
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
