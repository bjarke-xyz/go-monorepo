package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/generated"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/util"
)

// CreateRecipe is the resolver for the createRecipe field.
func (r *mutationResolver) CreateRecipe(ctx context.Context, input model.NewRecipe) (*model.Recipe, error) {
	gc, err := util.GinContextFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := util.GinContextUser(gc)
	if !ok {
		return nil, fmt.Errorf("must be logged in")
	}

	newRecipe := &model.Recipe{
		Title:          input.Title,
		Description:    input.Description,
		UserID:         user.ID,
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Published:      false,
	}
	createdRecipe, err := r.recipeRepository.CreateRecipe(newRecipe)
	return createdRecipe, err
}

// Recipes is the resolver for the recipes field.
func (r *queryResolver) Recipes(ctx context.Context) ([]*model.Recipe, error) {
	return r.recipeRepository.GetRecipes()
}

// User is the resolver for the user field.
func (r *recipeResolver) User(ctx context.Context, obj *model.Recipe) (*model.User, error) {
	log.Println("resolving user with id ", obj.UserID)
	user, err := r.userRepository.GetUser(obj.UserID)
	if err != nil {
		return nil, err
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
