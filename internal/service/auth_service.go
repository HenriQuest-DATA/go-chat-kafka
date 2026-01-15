package service

import (
	"chat-kafka-go/internal/config"
	"chat-kafka-go/internal/repository"
	"chat-kafka-go/pkg/types"
	"chat-kafka-go/pkg/utils"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthService gerencia autenticação e autorização
type AuthService struct {
	queries *repository.Queries // Repository gerado pelo SQLC
	cfg     *config.Config      // Configurações (JWT secrets, etc)
}

// NewAuthService cria nova instância do service
func NewAuthService(queries *repository.Queries, cfg *config.Config) *AuthService {
	return &AuthService{
		queries: queries,
		cfg:     cfg,
	}
}

// Register cria um novo usuário e retorna tokens
func (s *AuthService) Register(ctx context.Context, input types.RegisterInput) (*types.AuthResponse, error) {
	// 1. Validar input
	if err := s.validateRegisterInput(input); err != nil {
		return nil, err
	}

	// 2. Verificar se email já existe
	_, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err == nil {
		// Email encontrado = já existe
		return nil, fmt.Errorf("email já cadastrado")
	}
	if err != pgx.ErrNoRows {
		// Erro diferente de "não encontrado"
		return nil, fmt.Errorf("erro ao verificar email: %w", err)
	}

	// 3. Verificar se username já existe
	_, err = s.queries.GetUserByUsername(ctx, input.Username)
	if err == nil {
		return nil, fmt.Errorf("username já cadastrado")
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("erro ao verificar username: %w", err)
	}

	// 4. Hash da senha
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar hash da senha: %w", err)
	}

	// 5. Criar usuário no banco
	user, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar usuário: %w", err)
	}

	// 6. Gerar tokens JWT
	tokens, err := s.generateTokens(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar tokens: %w", err)
	}

	// 7. Salvar refresh token no banco
	if err := s.saveRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, fmt.Errorf("erro ao salvar refresh token: %w", err)
	}

	// 8. Montar resposta
	return &types.AuthResponse{
		User: &types.UserResponse{
			ID:        utils.UUIDToString(user.ID), // Converte UUID para string
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		},
		Tokens: tokens,
	}, nil
}

// validateRegisterInput valida dados de entrada
func (s *AuthService) validateRegisterInput(input types.RegisterInput) error {
	if input.Username == "" {
		return fmt.Errorf("username é obrigatório")
	}
	if len(input.Username) < 3 || len(input.Username) > 50 {
		return fmt.Errorf("username deve ter entre 3 e 50 caracteres")
	}

	if input.Email == "" {
		return fmt.Errorf("email é obrigatório")
	}
	// Validação básica de email (pode usar regex mais complexo)
	if !contains(input.Email, "@") || !contains(input.Email, ".") {
		return fmt.Errorf("email inválido")
	}

	if input.Password == "" {
		return fmt.Errorf("senha é obrigatória")
	}
	if len(input.Password) < 6 {
		return fmt.Errorf("senha deve ter no mínimo 6 caracteres")
	}

	return nil
}

// contains verifica se string contém substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		len(s) >= len(substr) && s != substr &&
		(s[0:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Login autentica usuário e retorna tokens
func (s *AuthService) Login(ctx context.Context, input types.LoginInput) (*types.AuthResponse, error) {
	// 1. Validar input
	if input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("email e senha são obrigatórios")
	}

	// 2. Buscar usuário por email
	user, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("credenciais inválidas")
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// 3. Verificar senha
	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// 4. Gerar novos tokens
	tokens, err := s.generateTokens(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar tokens: %w", err)
	}

	// 5. Salvar refresh token no banco
	if err := s.saveRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, fmt.Errorf("erro ao salvar refresh token: %w", err)
	}

	// 6. Retornar resposta
	return &types.AuthResponse{
		User: &types.UserResponse{
			ID:        utils.UUIDToString(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		},
		Tokens: tokens,
	}, nil
}

// RefreshToken renova access token usando refresh token válido
func (s *AuthService) RefreshToken(ctx context.Context, input types.RefreshTokenInput) (*types.TokenPair, error) {
	// 1. Validar input
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token é obrigatório")
	}

	// 2. Validar JWT do refresh token
	userID, err := utils.ValidateRefreshToken(input.RefreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("refresh token inválido: %w", err)
	}

	// 3. Verificar se refresh token existe no banco (não foi revogado)
	tokenRecord, err := s.queries.GetRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("refresh token inválido ou expirado")
		}
		return nil, fmt.Errorf("erro ao buscar refresh token: %w", err)
	}

	// 4. Buscar dados do usuário
	userUUID := pgtype.UUID{}
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("userID inválido: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// 5. Gerar novo access token (refresh token continua o mesmo)
	accessToken, err := utils.GenerateAccessToken(
		utils.UUIDToString(user.ID),
		user.Username,
		user.Email,
		s.cfg.JWT.AccessSecret,
		s.cfg.JWT.AccessExpiration,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar access token: %w", err)
	}

	// 6. Retornar novos tokens
	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: tokenRecord.Token, // Mesmo refresh token
	}, nil
}

// Logout invalida refresh token do usuário
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// 1. Validar input
	if refreshToken == "" {
		return fmt.Errorf("refresh token é obrigatório")
	}

	// 2. Deletar refresh token do banco (revoga)
	if err := s.queries.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("erro ao revogar token: %w", err)
	}

	return nil
}

// generateTokens gera access token e refresh token
func (s *AuthService) generateTokens(userID pgtype.UUID, username, email string) (*types.TokenPair, error) {
	// Access Token (1 hora)
	accessToken, err := utils.GenerateAccessToken(
		utils.UUIDToString(userID),
		username,
		email,
		s.cfg.JWT.AccessSecret,
		s.cfg.JWT.AccessExpiration,
	)
	if err != nil {
		return nil, err
	}

	// Refresh Token (7 dias)
	refreshToken, err := utils.GenerateRefreshToken(
		utils.UUIDToString(userID),
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.RefreshExpiration,
	)
	if err != nil {
		return nil, err
	}

	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// saveRefreshToken salva refresh token no banco
func (s *AuthService) saveRefreshToken(ctx context.Context, userID pgtype.UUID, token string) error {
	// Calcular expiração
	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(s.cfg.JWT.RefreshExpiration),
		Valid: true,
	}

	// Salvar no banco
	_, err := s.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})

	return err
}
