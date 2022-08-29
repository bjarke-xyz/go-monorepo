package middleware

import (
	"ShoppingList-Backend/internal/pkg/user"
	"ShoppingList-Backend/pkg/config"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type userContextKey string

func UserFromContext(ctx context.Context) *user.AppUser {
	appUser, ok := ctx.Value(userContextKey("user")).(*user.AppUser)
	if ok {
		return appUser
	}
	return nil
}

func SetContextUser(ctx context.Context, userId string) context.Context {
	appUser := &user.AppUser{ID: userId}
	newCtx := context.WithValue(ctx, userContextKey("user"), appUser)
	return newCtx
}

func JWTProtected(cfg *config.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zap.S()

			refreshInterval := time.Hour

			options := keyfunc.Options{
				RefreshInterval: &refreshInterval,
				RefreshErrorHandler: func(err error) {
					logger.Errorf("There was an error with the jwt.KeyFunc. Error: %v", err)
				},
			}

			jwks, err := keyfunc.Get(cfg.JwtJwksUrl, options)
			if err != nil {
				logger.Errorf("Failed to create JWKS from resource at the given URL. Error: %v", err)
				http.Error(w, "Could not create JWKS", http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				authHeaderQueryParam := r.URL.Query().Get("Authorization")
				authHeaderBytes, err := base64.StdEncoding.DecodeString(authHeaderQueryParam)
				if err == nil {
					authHeader = string(authHeaderBytes)
				}
			}
			if !strings.Contains(authHeader, "Bearer") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			jwtB64 := strings.TrimSpace(strings.Split(authHeader, "Bearer")[1])
			claims := jwt.StandardClaims{}
			token, err := jwt.ParseWithClaims(jwtB64, &claims, jwks.KeyFunc)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to parse the JWT. Error: %v", err)
				r.Header.Add("X-Error-Reason", errorMsg)
				http.Error(w, errorMsg, http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				r.Header.Add("X-Error-Reason", "Invalid token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := SetContextUser(r.Context(), claims.Subject)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
