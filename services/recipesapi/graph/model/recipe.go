package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Recipe struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    *string        `json:"description"`
	ImageID        *uuid.UUID     `json:"imageId"`
	Image          *Image         `json:"image"`
	UserID         string         `json:"userId"`
	User           *User          `json:"user"`
	CreatedAt      time.Time      `json:"createdDateTime"`
	ModeratedAt    *time.Time     `json:"moderatedDateTime"`
	LastModifiedAt time.Time      `json:"lastModifiedDateTime"`
	Published      bool           `json:"published"`
	Tips           []string       `json:"tips"`
	Yield          *string        `json:"yield"`
	Parts          []*RecipeParts `json:"parts"`
}

type RecipeParts struct {
	Title       string               `json:"title"`
	Ingredients []*RecipeIngredients `json:"ingredients"`
	Steps       []string             `json:"steps"`
}

type RecipeIngredients struct {
	Original string   `json:"original"`
	ID       *int     `json:"id"`
	Title    *string  `json:"title"`
	Volume   *float64 `json:"volume"`
	Unit     *string  `json:"unit"`
	Meta     []string `json:"meta"`
}

type RecipeRepository interface {
	GetRecipes(ctx context.Context) ([]*Recipe, error)
	GetRecipe(ctx context.Context, id string) (*Recipe, error)
	GetRecipeByTitle(ctx context.Context, title string) (*Recipe, error)
	SaveRecipe(ctx context.Context, recipe *Recipe) (*Recipe, error)
}

func MapRecipeInput(id string, input RecipeInput, user *User) *Recipe {
	newRecipe := &Recipe{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		UserID:      user.ID,
		Published:   input.Published,
		Tips:        input.Tips,
		Yield:       input.Yield,
		Parts:       make([]*RecipeParts, 0),
	}
	for _, part := range input.Parts {
		part := part
		recipePart := &RecipeParts{
			Title:       part.Title,
			Steps:       part.Steps,
			Ingredients: make([]*RecipeIngredients, 0),
		}
		for _, ingredient := range part.Ingredients {
			ingredient := ingredient
			recipePart.Ingredients = append(recipePart.Ingredients, &RecipeIngredients{
				Original: ingredient.Original,
				ID:       ingredient.ID,
				Title:    ingredient.Title,
				Volume:   ingredient.Volume,
				Unit:     ingredient.Unit,
				Meta:     ingredient.Meta,
			})
		}
		newRecipe.Parts = append(newRecipe.Parts, recipePart)
	}
	return newRecipe
}
