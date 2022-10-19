package model

import (
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/google/uuid"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserRepository struct {
	cfg   *config.Config
	users []*User
}

func NewUserRepository(cfg *config.Config) *UserRepository {
	repository := &UserRepository{
		cfg:   cfg,
		users: make([]*User, 0),
	}
	repository.CreateUser("test")
	return repository
}

func (u *UserRepository) GetUsers() ([]*User, error) {
	return u.users, nil
}

func (u *UserRepository) GetUser(id string) (*User, error) {
	return u.users[0], nil
	for _, user := range u.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (u *UserRepository) CreateUser(name string) (*User, error) {
	user := &User{
		ID:   uuid.NewString(),
		Name: name,
	}
	u.users = append(u.users, user)
	return user, nil
}
