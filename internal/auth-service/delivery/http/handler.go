package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/frishstrike/mercury-backend/internal/auth-service/domain"
	domainuc "github.com/frishstrike/mercury-backend/internal/auth-service/domain/usecase"
	"github.com/frishstrike/mercury-backend/pkg/auth"
)

type Handler struct {
	uc         domainuc.AuthUseCase
	jwtManager *auth.Manager
}

func NewHandler(uc domainuc.AuthUseCase, jwtManager *auth.Manager) *Handler {
	return &Handler{uc: uc, jwtManager: jwtManager}
}

// RegisterRoutes registers all auth routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/register", h.Register)
	mux.HandleFunc("/auth/login", h.Login)
	mux.HandleFunc("/auth/logout", h.Logout)
	mux.HandleFunc("/auth/refresh", h.Refresh)
	mux.HandleFunc("/auth/me", h.Me)
}

// Register godoc
// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, tokens, err := h.uc.Register(r.Context(), domainuc.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			writeError(w, http.StatusConflict, "user already exists")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user": map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
		"tokens": tokens,
	})
}

// Login godoc
// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, tokens, err := h.uc.Login(r.Context(), domainuc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		case domain.ErrUserInactive:
			writeError(w, http.StatusForbidden, "user is inactive")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
		"tokens": tokens,
	})
}

// Logout godoc
// POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	accessToken := extractBearerToken(r)
	if accessToken == "" {
		writeError(w, http.StatusUnauthorized, "missing access token")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims, err := h.jwtManager.ValidateToken(accessToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid access token")
		return
	}

	if err := h.uc.Logout(r.Context(), claims.UserID, accessToken, req.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "logged out successfully",
	})
}

// Refresh godoc
// POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	tokens, err := h.uc.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tokens": tokens,
	})
}

// Me godoc
// GET /auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	accessToken := extractBearerToken(r)
	if accessToken == "" {
		writeError(w, http.StatusUnauthorized, "missing access token")
		return
	}

	claims, err := h.jwtManager.ValidateToken(accessToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid access token")
		return
	}

	user, err := h.uc.GetUser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// helpers
func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}
