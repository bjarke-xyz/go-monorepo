package main

import (
	"fmt"
	"log"

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

	userRepository := model.NewUserRepository(cfg)
	recipeRepository := model.NewRecipeRepository(cfg)
	storageClient := storage.NewStorageClient(cfg)
	fileRepository := file.NewFileRepository(cfg)

	r := common.GinRouter(cfg)
	r.Use(util.GinContextToContextMiddleware())
	query := r.Group("/query", authMiddleware(cfg, userRepository))
	query.POST("", graphqlHandler(userRepository, recipeRepository, storageClient, fileRepository))
	r.GET("/", playgroundHandler())
	r.GET("/image/:id", imageHandler(storageClient, fileRepository))
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	r.Run()
}

func graphqlHandler(userRepository *model.UserRepository, recipeRepository *model.RecipeRepository, storage *storage.StorageClient, fileRepository *file.FileRepository) gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver.NewResolver(userRepository, recipeRepository, storage, fileRepository)}))
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

func imageHandler(storage *storage.StorageClient, fileRepository *file.FileRepository) gin.HandlerFunc {
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
		imageDto, err := fileRepository.GetFileById(imageUuid)
		if err != nil {
			ctx.AbortWithError(500, fmt.Errorf("failed to get image info: %w", err))
			return
		}
		if imageDto == nil {
			ctx.AbortWithError(404, fmt.Errorf("image not found"))
			return
		}
		bytes, err := storage.Get(imageDto.Bucket, imageDto.Key)
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
		if ctx.GetHeader("Authorization") == cfg.JobKey {
			user, _ := userRepository.GetUser("")
			ctx.Set("user", user)
		}
		ctx.Next()
	}
}
