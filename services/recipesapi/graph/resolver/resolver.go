package resolver

import (
	"github.com/bjarke-xyz/go-monorepo/libs/common/storage"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/file"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/recipes"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/users.go"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	recipeService *recipes.RecipeService
	userService   *users.UserService
	storage       *storage.StorageClient
	fileService   *file.FileService
}

func NewResolver(userService *users.UserService, recipeService *recipes.RecipeService, storage *storage.StorageClient, fileService *file.FileService) *Resolver {
	return &Resolver{
		recipeService: recipeService,
		userService:   userService,
		storage:       storage,
		fileService:   fileService,
	}
}
