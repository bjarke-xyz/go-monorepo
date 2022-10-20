package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

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
func (r *mutationResolver) SignUp(ctx context.Context, email string, password string) (*model.Token, error) {
	signInResp, err := r.userService.SignUp(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("could not sign up: %w", err)
	}
	if signInResp.Error != nil {
		return nil, fmt.Errorf("could not sign up: %v", signInResp.Error.Message)
	}
	return &model.Token{
		Token: signInResp.IdToken,
	}, nil
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
