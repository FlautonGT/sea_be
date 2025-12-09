package utils

import (
	"encoding/json"
	"net/http"

	"seaply/internal/domain"
)

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteSuccessJSON writes a successful JSON response
func WriteSuccessJSON(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, domain.SuccessResponse{Data: data})
}

// WriteCreatedJSON writes a 201 Created JSON response
func WriteCreatedJSON(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, domain.SuccessResponse{Data: data})
}

// WriteErrorJSON writes an error JSON response
func WriteErrorJSON(w http.ResponseWriter, status int, code, message, details string) {
	WriteJSON(w, status, domain.ErrorResponse{
		Error: domain.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// WriteValidationErrorJSON writes a validation error JSON response
func WriteValidationErrorJSON(w http.ResponseWriter, message string, fields map[string]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, domain.ErrorResponse{
		Error: domain.ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
	})
}

// WriteListJSON writes a paginated list response
func WriteListJSON(w http.ResponseWriter, data interface{}, pagination *domain.Pagination) {
	WriteJSON(w, http.StatusOK, domain.ListResponse{
		Data:       data,
		Pagination: pagination,
	})
}

// WriteAdminJSON writes an admin response with meta
func WriteAdminJSON(w http.ResponseWriter, data interface{}, permission, adminID, adminRole string) {
	WriteJSON(w, http.StatusOK, domain.AdminAPIResponse{
		Data: data,
		Meta: &domain.AdminMeta{
			RequiredPermission: permission,
			AdminID:            adminID,
			AdminRole:          adminRole,
		},
	})
}

// Common error responses
func WriteNotFoundError(w http.ResponseWriter, resource string) {
	WriteErrorJSON(w, http.StatusNotFound, "NOT_FOUND", resource+" not found", "")
}

func WriteUnauthorizedError(w http.ResponseWriter) {
	WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
}

func WriteForbiddenError(w http.ResponseWriter, permission string) {
	WriteErrorJSON(w, http.StatusForbidden, "PERMISSION_DENIED",
		"Anda tidak memiliki akses untuk melakukan aksi ini",
		"Required permission: "+permission)
}

func WriteInternalServerError(w http.ResponseWriter) {
	WriteErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR",
		"Terjadi kesalahan internal server", "")
}

func WriteBadRequestError(w http.ResponseWriter, message string) {
	WriteErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

func WriteConflictError(w http.ResponseWriter, message string) {
	WriteErrorJSON(w, http.StatusConflict, "CONFLICT", message, "")
}
