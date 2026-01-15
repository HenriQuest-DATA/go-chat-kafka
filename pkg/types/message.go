package types

// MessageResponse resposta de mensagem
type MessageResponse struct {
	ID         string `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// SendMessageInput dados para enviar mensagem
type SendMessageInput struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// ListMessagesInput dados para listar mensagens
type ListMessagesInput struct {
	UserID   string `json:"user_id"`
	FriendID string `json:"friend_id"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}
