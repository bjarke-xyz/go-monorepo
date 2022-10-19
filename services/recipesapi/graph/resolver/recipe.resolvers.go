package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/file"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/generated"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/util"
	"github.com/google/uuid"
)

// CreateRecipe is the resolver for the createRecipe field.
func (r *mutationResolver) CreateRecipe(ctx context.Context, input model.RecipeInput) (*model.Recipe, error) {
	user, ok := util.GinContextUser(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get user")
	}
	existingRecipe, err := r.recipeRepository.GetRecipeByTitle(input.Title)
	if err != nil {
		return nil, fmt.Errorf("error validating recipe title: %w", err)
	}
	if existingRecipe != nil {
		return nil, fmt.Errorf("recipe with title '%v' already exists", existingRecipe.Title)
	}
	recipeId := uuid.NewString()
	newRecipe := model.MapRecipeInput(recipeId, input, user)

	if input.Image != nil {
		if input.Image.Size > 5242880 {
			return nil, fmt.Errorf("image must be less than 5 megabytes")
		}
		imageData, err := io.ReadAll(input.Image.File)
		if err != nil {
			return nil, fmt.Errorf("could not read image file: %w", err)
		}
		imageId := uuid.New()
		key := fmt.Sprintf("/images/%v/%v", recipeId, imageId.String())
		err = r.storage.Put("recipesapi", key, imageData)
		if err != nil {
			return nil, fmt.Errorf("failed to store image: %w", err)
		}
		fileDto := &file.FileDto{
			ID:          imageId,
			Bucket:      "recipesapi",
			Key:         key,
			ContentType: input.Image.ContentType,
			Size:        input.Image.Size,
			Name:        input.Image.Filename,
		}
		err = r.fileRepository.SaveFile(fileDto)
		if err != nil {
			return nil, fmt.Errorf("failed to save image info: %w", err)
		}
		newRecipe.ImageID = &imageId
	}

	createdRecipe, err := r.recipeRepository.SaveRecipe(newRecipe)
	if err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}
	return createdRecipe, nil
}

// UpdateRecipe is the resolver for the updateRecipe field.
func (r *mutationResolver) UpdateRecipe(ctx context.Context, id string, input model.RecipeInput) (*model.Recipe, error) {
	user, ok := util.GinContextUser(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get user")
	}
	// TODO: handle image argument
	newRecipe := model.MapRecipeInput(id, input, user)
	createdRecipe, err := r.recipeRepository.SaveRecipe(newRecipe)
	if err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}
	return createdRecipe, nil
}

// Recipes is the resolver for the recipes field.
func (r *queryResolver) Recipes(ctx context.Context) ([]*model.Recipe, error) {
	recipes, err := r.recipeRepository.GetRecipes()
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}
	return recipes, nil
}

// GetRecipe is the resolver for the getRecipe field.
func (r *queryResolver) GetRecipe(ctx context.Context, id string) (*model.Recipe, error) {
	recipe, err := r.recipeRepository.GetRecipe(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe with id %v: %w", id, err)
	}
	return recipe, nil
}

// Image is the resolver for the image field.
func (r *recipeResolver) Image(ctx context.Context, obj *model.Recipe) (*model.Image, error) {
	if obj.ImageID == nil {
		return nil, nil
	}
	fileDto, err := r.fileRepository.GetFileById(*obj.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}
	if fileDto == nil {
		return nil, fmt.Errorf("no image info found")
	}
	return &model.Image{
		Src:  fmt.Sprintf("/image/%v", fileDto.ID.String()),
		Type: fileDto.ContentType,
		Size: int(fileDto.Size),
		Name: fileDto.Name,
	}, nil
}

// User is the resolver for the user field.
func (r *recipeResolver) User(ctx context.Context, obj *model.Recipe) (*model.User, error) {
	user, err := r.userRepository.GetUser(obj.UserID)
	if err != nil {
		return nil, fmt.Errorf("error with id '%v' not found: %w", obj.UserID, err)
	}
	return user, nil
}

// CreatedDateTime is the resolver for the createdDateTime field.
func (r *recipeResolver) CreatedDateTime(ctx context.Context, obj *model.Recipe) (*time.Time, error) {
	return &obj.CreatedAt, nil
}

// ModeratedDateTime is the resolver for the moderatedDateTime field.
func (r *recipeResolver) ModeratedDateTime(ctx context.Context, obj *model.Recipe) (*time.Time, error) {
	return obj.ModeratedAt, nil
}

// LastModifiedDateTime is the resolver for the lastModifiedDateTime field.
func (r *recipeResolver) LastModifiedDateTime(ctx context.Context, obj *model.Recipe) (*time.Time, error) {
	return &obj.LastModifiedAt, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Recipe returns generated.RecipeResolver implementation.
func (r *Resolver) Recipe() generated.RecipeResolver { return &recipeResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type recipeResolver struct{ *Resolver }
