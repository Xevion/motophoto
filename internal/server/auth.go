package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	validator "github.com/go-playground/validator/v10"
	nanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/middleware"
)

// validationMessage converts a validator.ValidationErrors into a single
// user-friendly string, hiding internal struct names and tag syntax.
func validationMessage(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return "invalid request"
	}
	e := ve[0]
	field := strings.ToLower(e.Field())
	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(e.Param(), " ", ", "))
	default:
		return field + " is invalid"
	}
}

func userResponseFromDB(u db.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        string(u.Role),
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, validationMessage(err))
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

	writeJSON(w, http.StatusOK, ItemResponse[UserResponse]{Data: userResponseFromDB(user)})
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
		writeError(w, http.StatusBadRequest, validationMessage(err))
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

	id, err := nanoid.New()
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("generating user id", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.LoggerFromContext(r.Context()).Error("hashing password", "error", err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		middleware.LoggerFromContext(r.Context()).Error("creating user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	if err := s.sessions.RenewToken(r.Context()); err != nil {
		middleware.LoggerFromContext(r.Context()).Error("renewing session token", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}
	s.sessions.Put(r.Context(), "user_id", user.ID)

	writeJSON(w, http.StatusCreated, ItemResponse[UserResponse]{Data: userResponseFromDB(user)})
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, ItemResponse[UserResponse]{Data: userResponseFromDB(*user)})
}
