package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToString converte pgtype.UUID para string
func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}

	// Converte [16]byte para uuid.UUID e depois para string
	return uuid.UUID(u.Bytes).String()
}

// StringToUUID converte string para pgtype.UUID
func StringToUUID(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("UUID inválido: %w", err)
	}

	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}, nil
}
