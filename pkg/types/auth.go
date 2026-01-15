package types

import "github.com/golang-jwt/jwt/v5"

// Claims estrutura customizada para JWT Access Token
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair par de tokens (access + refresh)
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse resposta completa de autenticação
type AuthResponse struct {
	User   *UserResponse `json:"user"`
	Tokens *TokenPair    `json:"tokens"`
}

// UserResponse dados públicos do usuário (sem password_hash)
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// RegisterInput dados necessários para registro
type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput dados necessários para login
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshTokenInput dados para refresh
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}
