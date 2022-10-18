package model

import (
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/google/uuid"
)

type Recipe struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    *string        `json:"description"`
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

type RecipeIngredients struct {
	Original string   `json:"original"`
	ID       *int     `json:"id"`
	Title    *string  `json:"title"`
	Volume   *float64 `json:"volume"`
	Unit     *string  `json:"unit"`
	Meta     []string `json:"meta"`
}

type RecipeParts struct {
	Title       string               `json:"title"`
	Ingredients []*RecipeIngredients `json:"ingredients"`
	Steps       []string             `json:"steps"`
}

type RecipeRepository struct {
	cfg     *config.Config
	recipes []*Recipe
}

func NewRecipeRepository(cfg *config.Config) *RecipeRepository {
	return &RecipeRepository{
		cfg:     cfg,
		recipes: make([]*Recipe, 0),
	}
}

func (r *RecipeRepository) GetRecipes() ([]*Recipe, error) {
	return r.recipes, nil
}

func (r *RecipeRepository) CreateRecipe(recipe *Recipe) (*Recipe, error) {
	recipe.ID = uuid.NewString()
	r.recipes = append(r.recipes, recipe)
	return recipe, nil
}
