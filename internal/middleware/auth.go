package middleware

import (
	"context"
	"net/http"
	"strings"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/jwtutil"
)

type middlewareContextKey string

const userIDKey middlewareContextKey = "userID"

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func JWTBearerMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			rawToken, isBearer := strings.CutPrefix(authHeader, "Bearer ")

			if !isBearer {
				apierr.InvalidAuth().RespondTo(w)
				return
			}

			id, err := jwtutil.Parse(rawToken, secret)
			if err != nil {
				apierr.InvalidAuth().RespondTo(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
