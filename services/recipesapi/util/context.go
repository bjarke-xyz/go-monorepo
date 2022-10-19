package util

import (
	"context"
	"fmt"

	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
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

func GinContextUser(ctx context.Context) (*model.User, bool) {
	gc, err := GinContextFromContext(ctx)
	if err != nil {
		return nil, false
	}
	userAny, ok := gc.Get("user")
	if !ok {
		return nil, false
	}

	user, ok := userAny.(*model.User)
	if !ok {
		return nil, false
	}
	return user, true
}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
