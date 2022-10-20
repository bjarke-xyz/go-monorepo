package util

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ginContextKeyType string

const ginContextKey ginContextKeyType = "GinContextKey"

func GinContextFromContext(ctx context.Context) (*gin.Context, error) {
	ginContext := ctx.Value(ginContextKey)
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return nil, err
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return nil, err
	}
	return gc, nil
}

func GinContextUserId(ctx context.Context) (string, bool) {
	gc, err := GinContextFromContext(ctx)
	if err != nil {
		return "", false
	}
	userAny, ok := gc.Get("userid")
	if !ok {
		return "", false
	}

	userId, ok := userAny.(string)
	if !ok {
		return "", false
	}
	return userId, true
}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
