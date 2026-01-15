package service

import (
	"context"
	"fmt"
	"time"

	"chat-kafka-go/internal/repository"
	"chat-kafka-go/pkg/types"
	"chat-kafka-go/pkg/utils"

	"github.com/jackc/pgx/v5"
)

// UserService gerencia operações de usuários
type UserService struct {
	queries *repository.Queries
}

// NewUserService cria nova instância do service
func NewUserService(queries *repository.Queries) *UserService {
	return &UserService{
		queries: queries,
	}
}

// GetUserByID busca usuário por ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*types.UserResponse, error) {
	// Converter string para UUID
	uuid, err := utils.StringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuário inválido: %w", err)
	}

	// Buscar no banco
	user, err := s.queries.GetUserByID(ctx, uuid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("usuário não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Retornar resposta (sem password_hash!)
	return &types.UserResponse{
		ID:        utils.UUIDToString(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

// GetUserByUsername busca usuário por username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*types.UserResponse, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("usuário não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return &types.UserResponse{
		ID:        utils.UUIDToString(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

// ListUsers lista usuários com paginação
func (s *UserService) ListUsers(ctx context.Context, input types.ListUsersInput) (*types.PaginatedResponse, error) {
	// Validar paginação
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 20 // Default: 20 por página
	}

	// Calcular offset
	offset := (input.Page - 1) * input.PerPage

	// Buscar usuários
	users, err := s.queries.ListUsers(ctx, repository.ListUsersParams{
		Limit:  int32(input.PerPage),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar usuários: %w", err)
	}

	// Converter para UserResponse (sem password_hash)
	userResponses := make([]types.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = types.UserResponse{
			ID:        utils.UUIDToString(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	// TODO: Buscar total de usuários para calcular totalPages
	// Por enquanto, vamos retornar meta básico
	return &types.PaginatedResponse{
		Success: true,
		Data:    userResponses,
		Meta: types.PaginationMeta{
			Page:       input.Page,
			PerPage:    input.PerPage,
			Total:      len(users), // Não é o total real, apenas da página
			TotalPages: 0,          // Calcular depois
		},
	}, nil
}

// AddFriend envia solicitação de amizade
func (s *UserService) AddFriend(ctx context.Context, input types.AddFriendInput) error {
	// Validar IDs
	if input.UserID == input.FriendID {
		return fmt.Errorf("não é possível adicionar a si mesmo como amigo")
	}

	// Converter UUIDs
	userUUID, err := utils.StringToUUID(input.UserID)
	if err != nil {
		return fmt.Errorf("ID de usuário inválido: %w", err)
	}

	friendUUID, err := utils.StringToUUID(input.FriendID)
	if err != nil {
		return fmt.Errorf("ID de amigo inválido: %w", err)
	}

	// Verificar se amizade já existe
	_, err = s.queries.GetFriendship(ctx, repository.GetFriendshipParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	})
	if err == nil {
		return fmt.Errorf("solicitação de amizade já existe")
	}
	if err != pgx.ErrNoRows {
		return fmt.Errorf("erro ao verificar amizade: %w", err)
	}

	// Criar solicitação de amizade (status: pending)
	_, err = s.queries.CreateFriendship(ctx, repository.CreateFriendshipParams{
		UserID:   userUUID,
		FriendID: friendUUID,
		Status:   "pending",
	})
	if err != nil {
		return fmt.Errorf("erro ao criar solicitação de amizade: %w", err)
	}

	return nil
}

// AcceptFriend aceita solicitação de amizade
func (s *UserService) AcceptFriend(ctx context.Context, input types.AcceptFriendInput) error {
	// Converter UUIDs
	userUUID, err := utils.StringToUUID(input.UserID)
	if err != nil {
		return fmt.Errorf("ID de usuário inválido: %w", err)
	}

	friendUUID, err := utils.StringToUUID(input.FriendID)
	if err != nil {
		return fmt.Errorf("ID de amigo inválido: %w", err)
	}

	// Buscar solicitação de amizade
	friendship, err := s.queries.GetFriendship(ctx, repository.GetFriendshipParams{
		UserID:   friendUUID, // Inverter: friend enviou para user
		FriendID: userUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("solicitação de amizade não encontrada")
		}
		return fmt.Errorf("erro ao buscar amizade: %w", err)
	}

	// Verificar se já está aceita
	if friendship.Status == "accepted" {
		return fmt.Errorf("amizade já aceita")
	}

	// Atualizar status para 'accepted'
	err = s.queries.UpdateFriendshipStatus(ctx, repository.UpdateFriendshipStatusParams{
		ID:     friendship.ID,
		Status: "accepted",
	})
	if err != nil {
		return fmt.Errorf("erro ao aceitar amizade: %w", err)
	}

	return nil
}

// ListFriends lista amigos aceitos de um usuário
func (s *UserService) ListFriends(ctx context.Context, userID string) ([]types.UserResponse, error) {
	// Converter UUID
	uuid, err := utils.StringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuário inválido: %w", err)
	}

	// Buscar amigos
	friends, err := s.queries.ListUserFriends(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar amigos: %w", err)
	}

	// Converter para UserResponse
	friendResponses := make([]types.UserResponse, len(friends))
	for i, friend := range friends {
		friendResponses[i] = types.UserResponse{
			ID:        utils.UUIDToString(friend.ID),
			Username:  friend.Username,
			Email:     friend.Email,
			CreatedAt: friend.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return friendResponses, nil
}
