package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/apperr"
	"github.com/thxhix/passKeeper/internal/domain/user"
	"github.com/thxhix/passKeeper/internal/mocks"
	"github.com/thxhix/passKeeper/internal/services"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper: create handlers with mocked dependencies
func makeHandlers(authSvc services.IAuthService) *Handlers {
	return &Handlers{
		authService: authSvc,
		logger:      zap.NewNop(),
	}
}

func TestHandlers_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"user1","password":"pass1"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		rec := httptest.NewRecorder()

		authSvc.On("Register", mock.Anything, "user1", "pass1").Return(int64(1), "access-token", "refresh-token", nil)

		h.Register(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.Equal(t, "no-store", res.Header.Get("Cache-Control"))

		b, _ := io.ReadAll(res.Body)
		expectedJSON := `{"access_token":"access-token", "id":1, "refresh_token":"refresh-token"}`
		assert.JSONEq(t, expectedJSON, string(b))

		authSvc.AssertExpectations(t)
	})

	t.Run("bad json", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		// invalid JSON
		body := `{"login": "u", "password": }`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("duplicate login -> conflict", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"dup","password":"pass"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		rec := httptest.NewRecorder()

		// service returns domain error DuplicateLogin
		authSvc.On("Register", mock.Anything, "dup", "pass").
			Return(int64(0), "", "", user.ErrDuplicateLogin)

		h.Register(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusConflict, res.StatusCode)
		authSvc.AssertExpectations(t)
	})

	t.Run("validation error -> bad request", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"bad","password":"p"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		rec := httptest.NewRecorder()

		// return a ValidationError (apperr.ValidationError) from service
		ve := &apperr.ValidationError{Message: "invalid"}
		authSvc.On("Register", mock.Anything, "bad", "p").Return(int64(0), "", "", ve)

		h.Register(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		authSvc.AssertExpectations(t)
	})
}

func TestHandlers_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"user","password":"pass"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		authSvc.On("Login", mock.Anything, "user", "pass").
			Return(int64(10), "access", "refresh", nil)

		h.Login(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		b, _ := io.ReadAll(res.Body)
		expected := `{"access_token":"access","refresh_token":"refresh","id":10}`
		assert.JSONEq(t, expected, string(b))

		authSvc.AssertExpectations(t)
	})

	t.Run("bad json", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"u", "password": }`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid credentials -> unauthorized", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"wrong","password":"pwd"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		// service returns apperr.AuthError (or any error that satisfies errors.As to apperr.AuthError)
		authErr := &apperr.AuthError{Message: "invalid"}
		authSvc.On("Login", mock.Anything, "wrong", "pwd").Return(int64(0), "", "", authErr)

		h.Login(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		authSvc.AssertExpectations(t)
	})

	t.Run("validation error -> bad request", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"login":"bad","password":"p"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		ve := &apperr.ValidationError{Message: "invalid"}
		authSvc.On("Login", mock.Anything, "bad", "p").Return(int64(0), "", "", ve)

		h.Login(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		authSvc.AssertExpectations(t)
	})
}

func TestHandlers_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		// choose a token that token.ValidateRefreshToken should accept;
		// in most implementations a non-empty string is fine. If your ValidateRefreshToken
		// is stricter, adjust this to a valid refresh token format.
		body := `{"refresh_token":"good-refresh-token"}`
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
		rec := httptest.NewRecorder()

		// handler calls token.ValidateRefreshToken first; assuming it passes, it then calls authService.Refresh
		authSvc.On("Refresh", mock.Anything, "good-refresh-token").
			Return("new-access", "new-refresh", nil)

		h.Refresh(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		b, _ := io.ReadAll(res.Body)
		expected := `{"access_token":"new-access","refresh_token":"new-refresh"}`
		assert.JSONEq(t, expected, string(b))

		authSvc.AssertExpectations(t)
	})

	t.Run("bad json", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"refresh_token": }`
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid refresh format -> bad request", func(t *testing.T) {
		// If token.ValidateRefreshToken rejects empty or malformed tokens, the handler will return 400.
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"refresh_token":""}` // empty token should be invalid
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("service auth error -> unauthorized", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"refresh_token":"valid-but-not-authorized"}`
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
		rec := httptest.NewRecorder()

		// handler will call token.ValidateRefreshToken (assume it passes). Then authService.Refresh returns an AuthError
		authErr := &apperr.AuthError{Message: "invalid refresh"}
		authSvc.On("Refresh", mock.Anything, "valid-but-not-authorized").
			Return("", "", authErr)

		h.Refresh(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		authSvc.AssertExpectations(t)
	})

	t.Run("service internal error -> internal server error", func(t *testing.T) {
		authSvc := new(mocks.AuthServiceMock)
		h := makeHandlers(authSvc)

		body := `{"refresh_token":"valid-but-server-error"}`
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
		rec := httptest.NewRecorder()

		authSvc.On("Refresh", mock.Anything, "valid-but-server-error").
			Return("", "", errors.New("db down"))

		h.Refresh(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		authSvc.AssertExpectations(t)
	})
}
