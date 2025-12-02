package handlers

import (
	"encoding/json"
	"net/http"
	"ticketbooth-backend/models"
)

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, errorCode string, message string) {
	JSON(w, status, models.ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// NotFound writes a 404 response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict writes a 409 response
func Conflict(w http.ResponseWriter, errorCode string, message string) {
	Error(w, http.StatusConflict, errorCode, message)
}

// BadRequest writes a 400 response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// InternalServerError writes a 500 response
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
