package types

// ListUsersInput parâmetros para listar usuários
type ListUsersInput struct {
	Page    int // Página atual (1, 2, 3...)
	PerPage int // Itens por página
}

// AddFriendInput dados para adicionar amigo
type AddFriendInput struct {
	UserID   string // Quem está enviando a solicitação
	FriendID string // Quem vai receber
}

// AcceptFriendInput dados para aceitar amizade
type AcceptFriendInput struct {
	UserID   string // Quem está aceitando
	FriendID string // Quem enviou a solicitação
}
