package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bjarke-xyz/go-monorepo/libs/common"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/go-monorepo/libs/common/storage"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/file"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/generated"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/resolver"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/recipes"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/users.go"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg, err := config.NewConfig()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	err = db.Migrate("up", cfg.ConnectionString())
	if err != nil {
		log.Printf("failed to migrate: %v", err)
	}

	cache := db.NewRedisCache(cfg)

	userRepository := model.NewUserRepository(cfg)
	userService := users.NewUserService(userRepository, cache)

	recipeRepository := model.NewNoSqlRecipeRepository(cfg)
	recipeService := recipes.NewRecipeService(recipeRepository, cache)

	storageClient := storage.NewStorageClient(cfg)
	fileRepository := file.NewFileRepository(cfg)
	fileService := file.NewFileService(fileRepository, storageClient)

	r := common.GinRouter(cfg)
	r.Use(util.GinContextToContextMiddleware())
	query := r.Group("/query", authMiddleware(cfg, userRepository))
	query.POST("", graphqlHandler(userService, recipeService, storageClient, fileService))
	r.GET("/", playgroundHandler())
	r.GET("/image/:id", imageHandler(storageClient, fileService))
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	r.Run()
}

func graphqlHandler(userService *users.UserService, recipeService *recipes.RecipeService, storage *storage.StorageClient, fileService *file.FileService) gin.HandlerFunc {
	c := generated.Config{Resolvers: resolver.NewResolver(userService, recipeService, storage, fileService)}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (res interface{}, err error) {
		if role == model.RoleAnon {
			return next(ctx)
		}

		_, ok := util.GinContextUserId(ctx)
		if !ok {
			return nil, fmt.Errorf("requires login")
		}
		// TODO: admin check without getting user each time
		return next(ctx)
	}
	h := handler.NewDefaultServer(generated.NewExecutableSchema(c))
	return func(ctx *gin.Context) {
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL Playground", "/query")
	return func(ctx *gin.Context) {
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func imageHandler(storage *storage.StorageClient, fileService *file.FileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		imageId := ctx.Param("id")
		if imageId == "" {
			ctx.AbortWithError(400, nil)
			return
		}
		imageUuid, err := uuid.Parse(imageId)
		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}
		imageDto, err := fileService.GetFileById(ctx.Request.Context(), imageUuid)
		if err != nil {
			ctx.AbortWithError(500, fmt.Errorf("failed to get image info: %w", err))
			return
		}
		if imageDto == nil {
			ctx.AbortWithError(404, fmt.Errorf("image not found"))
			return
		}
		bytes, err := storage.Get(ctx.Request.Context(), imageDto.Bucket, imageDto.Key)
		if err != nil {
			ctx.AbortWithError(500, fmt.Errorf("failed to get image: %w", err))
			return
		}
		ctx.Header("Cache-Control", "public, max-age=86400")
		ctx.Data(200, imageDto.ContentType, bytes)
	}
}

func authMiddleware(cfg *config.Config, userRepository *model.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			authHeaderParts := strings.Split(authHeader, " ")
			if len(authHeaderParts) >= 2 {
				authHeader = strings.TrimSpace(authHeaderParts[1])
			}
		}
		if authHeader != "" {
			userId, err := userRepository.GetUserIdFromToken(ctx.Request.Context(), authHeader)
			if err != nil {
				log.Printf("failed to get user id: %v", err)
				ctx.AbortWithStatusJSON(400, gin.H{
					"errors": []gin.H{{
						"message": "failed to get user id",
						"path":    []string{},
					}},
					"data": nil,
				})
				return
			}
			ctx.Set("userid", userId)
		}
		ctx.Next()
	}
}
