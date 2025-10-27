package middleware

import (
	"context"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"net/http"
	"strconv"
	"strings"
)

type TokenParser interface {
	ParseAccessToken(tokenStr string) (userID string, err error)
}

type HTTPErrorResponser interface {
	PublicError(w http.ResponseWriter, code int, err error)
	InternalError(w http.ResponseWriter, err error)
}

type ctxKey string

const CtxUserID ctxKey = "user_id"

func Authorize(jwtManager TokenParser, errorResponser HTTPErrorResponser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				errorResponser.PublicError(w, http.StatusUnauthorized, token.ErrMissingBearerToken)
				return
			}

			providedToken := strings.TrimPrefix(h, "Bearer ")

			userID, err := jwtManager.ParseAccessToken(providedToken)
			if err != nil {
				errorResponser.PublicError(w, http.StatusUnauthorized, token.ErrInvalidAuthToken)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromCtx(ctx context.Context) (int64, bool) {
	v := ctx.Value(CtxUserID)
	str, ok := v.(string)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
