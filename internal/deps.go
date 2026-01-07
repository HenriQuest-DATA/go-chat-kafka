package internal

import (
	_ "github.com/IBM/sarama"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	_ "github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "golang.org/x/crypto/bcrypt"
)
