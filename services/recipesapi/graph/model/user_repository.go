package model

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/goccy/go-json"
)

type UserRepository struct {
	cfg    *config.Config
	app    *firebase.App
	client *auth.Client
}

func NewUserRepository(cfg *config.Config) *UserRepository {
	repository := &UserRepository{cfg: cfg}
	return repository
}

func (u *UserRepository) getClient(ctx context.Context) (*auth.Client, error) {
	if u.app == nil {
		app, err := firebase.NewApp(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get firebase app: %w", err)
		}
		u.app = app
	}
	if u.app == nil {
		return nil, fmt.Errorf("firebase app was nil")
	}

	if u.client == nil {
		client, err := u.app.Auth(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth client: %w", err)
		}
		u.client = client
	}
	if u.client == nil {
		return nil, fmt.Errorf("firestore auth was nil")
	}

	return u.client, nil
}

func (u *UserRepository) GetUserIdFromToken(ctx context.Context, idToken string) (string, error) {
	client, err := u.getClient(ctx)
	if err != nil {
		return "", err
	}
	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}
	return token.UID, nil
}

func (u *UserRepository) GetUserById(ctx context.Context, userId string) (*User, error) {
	client, err := u.getClient(ctx)
	if err != nil {
		return nil, err
	}
	user, err := client.GetUser(ctx, userId)
	if err != nil {
		if auth.IsUserNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("could not get user with id %v: %w", userId, err)
		}
	}
	return &User{
		ID:            user.UID,
		DisplayName:   user.DisplayName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}, nil
}

func (u *UserRepository) SignUp(ctx context.Context, email string, password string, displayName string) (*UserWithToken, error) {
	client, err := u.getClient(ctx)
	if err != nil {
		return nil, err
	}
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(displayName).
		Disabled(false)
	user, err := client.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	signInResp, err := u.SignIn(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	return &UserWithToken{
		User: User{
			ID:            user.UID,
			DisplayName:   user.DisplayName,
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		},
		Token: signInResp.IdToken,
	}, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, id string, email string, password *string, displayName string) (*User, error) {
	client, err := u.getClient(ctx)
	if err != nil {
		return nil, err
	}
	params := (&auth.UserToUpdate{}).Email(email).DisplayName(displayName)
	if password != nil {
		params = params.Password(*password)
	}
	user, err := client.UpdateUser(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &User{
		ID:            user.UID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		DisplayName:   user.DisplayName,
	}, nil
}

type RestApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SignInResponse struct {
	IdToken string        `json:"idToken"`
	Email   string        `json:"email"`
	LocalId string        `json:"localId"`
	Error   *RestApiError `json:"error,omitempty"`
}

func (u *UserRepository) SignIn(ctx context.Context, email string, password string) (*SignInResponse, error) {
	body := make(map[string]any)
	body["email"] = email
	body["password"] = password
	body["returnSecureToken"] = true
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="+u.cfg.FirebaseWebApiKey, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var signInResponse SignInResponse
	err = json.Unmarshal(respBody, &signInResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &signInResponse, nil
}
