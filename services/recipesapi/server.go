package main

import (
	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bjarke-xyz/go-monorepo/libs/common"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/generated"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/util"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg, err := config.NewConfig()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	userRepository := model.NewUserRepository(cfg)
	recipeRepository := model.NewRecipeRepository(cfg)

	r := common.GinRouter(cfg)
	r.Use(util.GinContextToContextMiddleware())
	query := r.Group("/query", authMiddleware(cfg, userRepository))
	query.POST("", graphqlHandler(userRepository, recipeRepository))
	r.GET("/", playgroundHandler())
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	r.Run()
}

func graphqlHandler(userRepository *model.UserRepository, recipeRepository *model.RecipeRepository) gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(userRepository, recipeRepository)}))
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

func authMiddleware(cfg *config.Config, userRepository *model.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Authorization") == cfg.JobKey {
			users, _ := userRepository.GetUsers()
			ctx.Set("user", users[0])
		}
		ctx.Next()
	}
}
