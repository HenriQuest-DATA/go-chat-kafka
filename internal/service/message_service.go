package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat-kafka-go/internal/repository"
	"chat-kafka-go/pkg/types"
	"chat-kafka-go/pkg/utils"
)

// MessageService gerencia mensagens
type MessageService struct {
	queries  *repository.Queries
	producer KafkaProducer // Interface para Kafka Producer
}

// KafkaProducer interface para enviar mensagens ao Kafka
// Vamos implementar depois, por enquanto é uma interface
type KafkaProducer interface {
	SendMessage(topic string, key string, value []byte) error
}

// NewMessageService cria nova instância do service
func NewMessageService(queries *repository.Queries, producer KafkaProducer) *MessageService {
	return &MessageService{
		queries:  queries,
		producer: producer,
	}
}

// SendMessage envia mensagem (salva no DB + envia para Kafka)
func (s *MessageService) SendMessage(ctx context.Context, input types.SendMessageInput) (*types.MessageResponse, error) {
	// 1. Validar input
	if err := s.validateSendMessageInput(input); err != nil {
		return nil, err
	}

	// 2. Converter UUIDs
	senderUUID, err := utils.StringToUUID(input.SenderID)
	if err != nil {
		return nil, fmt.Errorf("sender_id inválido: %w", err)
	}

	receiverUUID, err := utils.StringToUUID(input.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("receiver_id inválido: %w", err)
	}

	// 3. Salvar mensagem no banco com status 'sent'
	message, err := s.queries.CreateMessage(ctx, repository.CreateMessageParams{
		SenderID:   senderUUID,
		ReceiverID: receiverUUID,
		Content:    input.Content,
		Status:     "sent",
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar mensagem: %w", err)
	}

	// 4. Preparar mensagem para Kafka
	kafkaMessage := map[string]interface{}{
		"id":          utils.UUIDToString(message.ID),
		"sender_id":   input.SenderID,
		"receiver_id": input.ReceiverID,
		"content":     input.Content,
		"timestamp":   message.CreatedAt.Time.Unix(),
	}

	messageBytes, err := json.Marshal(kafkaMessage)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// 5. Enviar para Kafka (assíncrono)
	// Se producer for nil (testes), pula esta etapa
	if s.producer != nil {
		if err := s.producer.SendMessage("chat-messages", input.ReceiverID, messageBytes); err != nil {
			// Log erro mas não falha (mensagem já está no DB)
			fmt.Printf("WARN: Erro ao enviar para Kafka: %v\n", err)
		}
	}

	// 6. Retornar resposta
	return &types.MessageResponse{
		ID:         utils.UUIDToString(message.ID),
		SenderID:   utils.UUIDToString(message.SenderID),
		ReceiverID: utils.UUIDToString(message.ReceiverID),
		Content:    message.Content,
		Status:     message.Status,
		CreatedAt:  message.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

// validateSendMessageInput valida dados de entrada
func (s *MessageService) validateSendMessageInput(input types.SendMessageInput) error {
	if input.SenderID == "" {
		return fmt.Errorf("sender_id é obrigatório")
	}
	if input.ReceiverID == "" {
		return fmt.Errorf("receiver_id é obrigatório")
	}
	if input.SenderID == input.ReceiverID {
		return fmt.Errorf("não é possível enviar mensagem para si mesmo")
	}
	if input.Content == "" {
		return fmt.Errorf("conteúdo da mensagem é obrigatório")
	}
	if len(input.Content) > 5000 {
		return fmt.Errorf("mensagem muito longa (máximo 5000 caracteres)")
	}
	return nil
}

// GetMessagesBetween lista mensagens entre dois usuários
func (s *MessageService) GetMessagesBetween(ctx context.Context, input types.ListMessagesInput) (*types.PaginatedResponse, error) {
	// Validar paginação
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 50 // Default: 50 mensagens por página
	}

	// Converter UUIDs
	userUUID, err := utils.StringToUUID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user_id inválido: %w", err)
	}

	friendUUID, err := utils.StringToUUID(input.FriendID)
	if err != nil {
		return nil, fmt.Errorf("friend_id inválido: %w", err)
	}

	// Calcular offset
	offset := (input.Page - 1) * input.PerPage

	// Buscar mensagens
	messages, err := s.queries.ListMessagesBetweenUsers(ctx, repository.ListMessagesBetweenUsersParams{
		SenderID:   userUUID,
		ReceiverID: friendUUID,
		Limit:      int32(input.PerPage),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar mensagens: %w", err)
	}

	// Converter para MessageResponse
	messageResponses := make([]types.MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = types.MessageResponse{
			ID:         utils.UUIDToString(msg.ID),
			SenderID:   utils.UUIDToString(msg.SenderID),
			ReceiverID: utils.UUIDToString(msg.ReceiverID),
			Content:    msg.Content,
			Status:     msg.Status,
			CreatedAt:  msg.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return &types.PaginatedResponse{
		Success: true,
		Data:    messageResponses,
		Meta: types.PaginationMeta{
			Page:       input.Page,
			PerPage:    input.PerPage,
			Total:      len(messages),
			TotalPages: 0, // Calcular depois
		},
	}, nil
}

// MarkAsDelivered marca mensagem como entregue
func (s *MessageService) MarkAsDelivered(ctx context.Context, messageID string) error {
	uuid, err := utils.StringToUUID(messageID)
	if err != nil {
		return fmt.Errorf("message_id inválido: %w", err)
	}

	err = s.queries.UpdateMessageStatus(ctx, repository.UpdateMessageStatusParams{
		ID:     uuid,
		Status: "delivered",
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	return nil
}

// MarkAsRead marca mensagem como lida
func (s *MessageService) MarkAsRead(ctx context.Context, messageID string) error {
	uuid, err := utils.StringToUUID(messageID)
	if err != nil {
		return fmt.Errorf("message_id inválido: %w", err)
	}

	err = s.queries.UpdateMessageStatus(ctx, repository.UpdateMessageStatusParams{
		ID:     uuid,
		Status: "read",
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	return nil
}
