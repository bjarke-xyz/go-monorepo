package resolver

import (
	"github.com/bjarke-xyz/go-monorepo/libs/common/storage"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/file"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	recipeRepository model.RecipeRepository
	userRepository   *model.UserRepository
	storage          *storage.StorageClient
	fileService      *file.FileService
}

func NewResolver(userRepository *model.UserRepository, recipeRepository model.RecipeRepository, storage *storage.StorageClient, fileService *file.FileService) *Resolver {
	return &Resolver{
		recipeRepository: recipeRepository,
		userRepository:   userRepository,
		storage:          storage,
		fileService:      fileService,
	}
}
