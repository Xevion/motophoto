package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/server"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

func TestHandleLogin(t *testing.T) {
	t.Parallel()

	t.Run("valid credentials", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)
		ctx := t.Context()

		email := "login@example.com"
		password := "securepassword"
		dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{
			Email:    &email,
			Password: &password,
		})

		body := `{"email":"login@example.com","password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.UserResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "login@example.com", resp.Data.Email)
		assert.NotEmpty(t, resp.Data.ID)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)
		ctx := t.Context()

		email := "wrong-pw@example.com"
		password := "securepassword"
		dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{
			Email:    &email,
			Password: &password,
		})

		body := `{"email":"wrong-pw@example.com","password":"wrongpassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("email not found", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"nobody@example.com","password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("generic error message on failure", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"nobody@example.com","password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.NotContains(t, rr.Body.String(), "email")
		assert.NotContains(t, rr.Body.String(), "not found")
		assert.Contains(t, rr.Body.String(), "invalid credentials")
	})

	t.Run("banned user", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)
		ctx := t.Context()

		email := "banned@example.com"
		password := "securepassword"
		dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{
			Email:    &email,
			Password: &password,
			Banned:   true,
		})

		body := `{"email":"banned@example.com","password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("response does not contain password", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)
		ctx := t.Context()

		email := "nopw@example.com"
		password := "securepassword"
		dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{
			Email:    &email,
			Password: &password,
		})

		body := `{"email":"nopw@example.com","password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotContains(t, rr.Body.String(), "password")
		assert.NotContains(t, rr.Body.String(), "hash")
	})

	t.Run("missing email", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"password":"securepassword"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing password", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"test@example.com"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/login", `{bad`)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleLogout(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/logout", "")

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "logged out")
	})
}

func TestHandleRegister(t *testing.T) {
	t.Parallel()

	t.Run("valid photographer", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"jane@example.com","password":"securepassword","display_name":"Jane Doe","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp server.ItemResponse[server.UserResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "jane@example.com", resp.Data.Email)
		assert.Equal(t, "Jane Doe", resp.Data.DisplayName)
		assert.Equal(t, "photographer", resp.Data.Role)
		assert.NotEmpty(t, resp.Data.ID)
	})

	t.Run("valid customer", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"john@example.com","password":"securepassword","display_name":"John Smith","role":"customer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp server.ItemResponse[server.UserResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "customer", resp.Data.Role)
	})

	t.Run("response does not contain password", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"secret@example.com","password":"securepassword","display_name":"Secret","role":"customer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotContains(t, rr.Body.String(), "password")
		assert.NotContains(t, rr.Body.String(), "hash")
	})

	t.Run("missing email", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"password":"securepassword","display_name":"Jane","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid email format", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"not-an-email","password":"securepassword","display_name":"Jane","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("password too short", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"jane@example.com","password":"short","display_name":"Jane","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing display name", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"jane@example.com","password":"securepassword","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid role", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		body := `{"email":"jane@example.com","password":"securepassword","display_name":"Jane","role":"admin"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("duplicate email", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)
		ctx := t.Context()

		email := "existing@example.com"
		dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{Email: &email})

		body := `{"email":"existing@example.com","password":"securepassword","display_name":"Jane","role":"photographer"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		env := testutil.NewEnv(t)

		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/auth/register", `{bad`)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
