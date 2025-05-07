package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/userapi/internal/application"
	"github.com/yourusername/userapi/internal/domain"
	"github.com/yourusername/userapi/pkg/auth"
	"github.com/yourusername/userapi/pkg/validation"
)

// Handler holds services needed for HTTP handlers
type Handler struct {
	userService *application.UserService
	authService *application.AuthService
	jwtAuth     *auth.JWTAuth
}

// NewHandler creates a new HTTP handler
func NewHandler(userService *application.UserService, authService *application.AuthService, jwtAuth *auth.JWTAuth) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
		jwtAuth:     jwtAuth,
	}
}

// Response represents the standard API response format
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterHandler handles user registration
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if err := validation.Validate(input); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authService.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrEmailAlreadyExists {
			status = http.StatusConflict
		}
		respondWithError(w, status, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, Response{Success: true, Data: user})
}

// LoginHandler handles user authentication
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if err := validation.Validate(input); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrInvalidCredentials {
			status = http.StatusUnauthorized
		}
		respondWithError(w, status, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"token": token,
		},
	})
}

// GetUserHandler retrieves a user by ID
func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrUserNotFound {
			status = http.StatusNotFound
		}
		respondWithError(w, status, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{Success: true, Data: user})
}

// GetAllUsersHandler retrieves all users
func (h *Handler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{Success: true, Data: users})
}

// UpdateUserHandler updates a user
func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	var input struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if err := validation.Validate(input); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, input.Name, input.Email)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrUserNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrEmailAlreadyExists {
			status = http.StatusConflict
		}
		respondWithError(w, status, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{Success: true, Data: user})
}

// DeleteUserHandler deletes a user
func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	err := h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrUserNotFound {
			status = http.StatusNotFound
		}
		respondWithError(w, status, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{Success: true})
}

// Helper functions for HTTP responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, Response{Success: false, Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
