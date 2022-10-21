package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/generated"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/util"
)

// SignIn is the resolver for the signIn field.
func (r *mutationResolver) SignIn(ctx context.Context, email string, password string) (*model.Token, error) {
	signInResp, err := r.userService.SignIn(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("could not sign in: %w", err)
	}
	if signInResp.Error != nil {
		return nil, fmt.Errorf("could not sign in: %v", signInResp.Error.Message)
	}
	return &model.Token{
		Token: signInResp.IdToken,
	}, nil
}

// SignUp is the resolver for the signUp field.
func (r *mutationResolver) SignUp(ctx context.Context, input model.UserInput) (*model.Token, error) {
	if input.Password == nil {
		return nil, fmt.Errorf("invalid password")
	}
	signInResp, err := r.userService.SignUp(ctx, input.Email, *input.Password, input.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("could not sign up: %w", err)
	}
	return &model.Token{
		Token: signInResp.Token,
	}, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	userId, ok := util.GinContextUserId(ctx)
	if !ok {
		return nil, fmt.Errorf("requires login")
	}
	user, err := r.userService.UpdateUser(ctx, userId, input.Email, input.Password, input.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return user, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	userId, ok := util.GinContextUserId(ctx)
	if !ok {
		return nil, fmt.Errorf("requires login")
	}
	user, err := r.userService.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Recipes is the resolver for the recipes field.
func (r *userResolver) Recipes(ctx context.Context, obj *model.User) ([]*model.Recipe, error) {
	recipes, err := r.recipeService.GetRecipesByUserId(ctx, obj.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes for user: %w", err)
	}
	return recipes, nil
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *userResolver) DisplayName(ctx context.Context, obj *model.User) (string, error) {
	return obj.DisplayName, nil
}
