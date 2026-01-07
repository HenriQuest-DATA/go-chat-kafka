package utils

import (
	"chat-kafka-go/pkg/types"
	"encoding/json"
	"net/http"
)

// JSON envia resposta JSON genérica
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Se falhar ao encodar, loga o erro
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Success envia resposta de sucesso
func Success(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	JSON(w, statusCode, types.SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Error envia resposta de erro
func Error(w http.ResponseWriter, statusCode int, message string, code string) {
	JSON(w, statusCode, types.ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
	})
}
