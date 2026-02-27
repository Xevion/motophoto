package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	nanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/middleware"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err))
		return
	}

	user, err := s.queries.GetUserByEmail(r.Context(), req.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("looking up user by email", "error", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	if user.BannedAt.Valid {
		writeError(w, http.StatusForbidden, "account is banned")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := s.sessions.RenewToken(r.Context()); err != nil {
		middleware.LoggerFromContext(r.Context()).Error("renewing session token", "error", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	s.sessions.Put(r.Context(), "user_id", user.ID)

	writeJSON(w, http.StatusOK, ItemResponse[UserResponse]{Data: UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
	}})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Destroy(r.Context()); err != nil {
		middleware.LoggerFromContext(r.Context()).Error("destroying session", "error", err)
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	writeJSON(w, http.StatusOK, ItemResponse[map[string]string]{Data: map[string]string{"message": "logged out"}})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err))
		return
	}

	var role db.UserRole
	switch req.Role {
	case "photographer":
		role = db.UserRolePhotographer
	case "customer":
		role = db.UserRoleCustomer
	default:
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}

	_, err := s.queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		middleware.LoggerFromContext(r.Context()).Error("checking email uniqueness", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	id, err := nanoid.New()
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("generating user id", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("hashing password", "error", fmt.Errorf("hashing password: %w", err))
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	user, err := s.queries.CreateUser(r.Context(), db.CreateUserParams{
		ID:           id,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		Role:         role,
	})
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("creating user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	s.sessions.Put(r.Context(), "user_id", user.ID)

	writeJSON(w, http.StatusCreated, ItemResponse[UserResponse]{Data: UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
	}})
}
