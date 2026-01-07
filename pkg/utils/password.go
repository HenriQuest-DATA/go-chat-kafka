package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword gera hash bcrypt de uma senha
// Cost 12 = 2^12 iterações (+-250ms)
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("falha ao gerar hash: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifica se a senha bate com o hash
// Retorna true se a senha está correta
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
