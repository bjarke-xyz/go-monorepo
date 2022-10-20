package users

import (
	"context"
	"fmt"
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
)

type UserService struct {
	userRepository *model.UserRepository
	cache          *db.RedisCache
}

func NewUserService(userRepository *model.UserRepository, cache *db.RedisCache) *UserService {
	return &UserService{
		userRepository: userRepository,
		cache:          cache,
	}
}

func (u *UserService) GetUserIdFromToken(ctx context.Context, idToken string) (string, error) {
	return u.userRepository.GetUserIdFromToken(ctx, idToken)
}

func (u *UserService) GetUserById(ctx context.Context, userId string) (*model.User, error) {
	cacheKey := fmt.Sprintf("GetUserById:%v", userId)
	var user *model.User
	if err := u.cache.Get(ctx, cacheKey, &user); err == nil {
		return user, nil
	}
	user, err := u.userRepository.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	u.cache.Set(ctx, cacheKey, user, time.Minute*5)
	return user, nil
}

func (u *UserService) SignIn(ctx context.Context, email string, password string) (*model.SignInResponse, error) {
	return u.userRepository.SignIn(ctx, email, password)
}
func (u *UserService) SignUp(ctx context.Context, email string, password string) (*model.SignInResponse, error) {
	return u.userRepository.SignUp(ctx, email, password)
}
