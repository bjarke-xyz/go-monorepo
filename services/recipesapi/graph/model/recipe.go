package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Recipe struct {
	ID             string         `json:"id" db:"id"`
	Title          string         `json:"title" db:"title"`
	Description    *string        `json:"description" db:"description"`
	ImageID        *uuid.UUID     `json:"-"`
	Image          *Image         `json:"image" db:"image"`
	UserID         string         `json:"userId" db:"user_id"`
	User           *User          `json:"user"`
	CreatedAt      time.Time      `json:"createdDateTime" db:"created_at"`
	ModeratedAt    *time.Time     `json:"moderatedDateTime" db:"moderated_at"`
	LastModifiedAt time.Time      `json:"lastModifiedDateTime" db:"last_modified_at"`
	Published      bool           `json:"published" db:"published"`
	Tips           []string       `json:"tips" db:"tips"`
	Yield          *string        `json:"yield" db:"yield"`
	Parts          []*RecipeParts `json:"parts" db:"parts"`
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

type StringSlice []string
type PartsSlice []*recipePartsDto
type IngredientsSlice []*RecipeIngredients
type recipeDto struct {
	ID             string      `db:"id"`
	Title          string      `db:"title"`
	Description    *string     `db:"description"`
	ImageId        *uuid.UUID  `db:"image_id"`
	UserID         string      `db:"user_id"`
	CreatedAt      time.Time   `db:"created_at"`
	ModeratedAt    *time.Time  `db:"moderated_at"`
	LastModifiedAt time.Time   `db:"last_modified_at"`
	Published      bool        `db:"published"`
	Tips           StringSlice `db:"tips"`
	Yield          *string     `db:"yield"`
	Parts          PartsSlice  `json:"parts" db:"parts"`
}
type recipePartsDto struct {
	Title       string
	Ingredients IngredientsSlice
	Steps       []string
}

func (i *Image) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &i)
		return nil
	case string:
		json.Unmarshal([]byte(v), &i)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (i Image) Value() (driver.Value, error) {
	return json.Marshal(&i)
}

func (ss *StringSlice) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &ss)
		return nil
	case string:
		json.Unmarshal([]byte(v), &ss)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (ss StringSlice) Value() (driver.Value, error) {
	return json.Marshal(&ss)
}

func (ri *IngredientsSlice) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &ri)
		return nil
	case string:
		json.Unmarshal([]byte(v), &ri)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (ri IngredientsSlice) Value() (driver.Value, error) {
	return json.Marshal(&ri)
}

func (rp *PartsSlice) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &rp)
		return nil
	case string:
		json.Unmarshal([]byte(v), &rp)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (rp PartsSlice) Value() (driver.Value, error) {
	return json.Marshal(&rp)
}

type RecipeRepository struct {
	cfg *config.Config
}

func NewRecipeRepository(cfg *config.Config) *RecipeRepository {
	return &RecipeRepository{
		cfg: cfg,
	}
}

func mapRecipeDto(dto *recipeDto) *Recipe {
	recipe := &Recipe{
		ID:             dto.ID,
		Title:          dto.Title,
		Description:    dto.Description,
		ImageID:        dto.ImageId,
		UserID:         dto.UserID,
		CreatedAt:      dto.CreatedAt,
		ModeratedAt:    dto.ModeratedAt,
		LastModifiedAt: dto.LastModifiedAt,
		Published:      dto.Published,
		Tips:           dto.Tips,
		Yield:          dto.Yield,
		Parts:          make([]*RecipeParts, 0),
	}
	for _, dtoPart := range dto.Parts {
		dtoPart := dtoPart
		recipeParts := &RecipeParts{
			Title:       dtoPart.Title,
			Steps:       dtoPart.Steps,
			Ingredients: make([]*RecipeIngredients, 0),
		}
		for _, dtoIngredient := range dtoPart.Ingredients {
			dtoIngredient := dtoIngredient
			recipeParts.Ingredients = append(recipeParts.Ingredients, &RecipeIngredients{
				Original: dtoIngredient.Original,
				ID:       dtoIngredient.ID,
				Title:    dtoIngredient.Title,
				Volume:   dtoIngredient.Volume,
				Unit:     dtoIngredient.Unit,
				Meta:     dtoIngredient.Meta,
			})
		}
		recipe.Parts = append(recipe.Parts, recipeParts)
	}
	return recipe
}

func mapRecipe(recipe *Recipe) *recipeDto {
	dto := &recipeDto{
		ID:             recipe.ID,
		Title:          recipe.Title,
		Description:    recipe.Description,
		ImageId:        recipe.ImageID,
		UserID:         recipe.UserID,
		CreatedAt:      recipe.CreatedAt,
		ModeratedAt:    recipe.ModeratedAt,
		LastModifiedAt: recipe.LastModifiedAt,
		Published:      recipe.Published,
		Tips:           recipe.Tips,
		Yield:          recipe.Yield,
		Parts:          make(PartsSlice, 0),
	}
	for _, recipePart := range recipe.Parts {
		recipePart := recipePart
		dtoPart := &recipePartsDto{
			Title:       recipePart.Title,
			Steps:       recipePart.Steps,
			Ingredients: make(IngredientsSlice, 0),
		}
		for _, recipeIngredient := range recipePart.Ingredients {
			recipeIngredient := recipeIngredient
			dtoPart.Ingredients = append(dtoPart.Ingredients, &RecipeIngredients{
				Original: recipeIngredient.Original,
				ID:       recipeIngredient.ID,
				Title:    recipeIngredient.Title,
				Volume:   recipeIngredient.Volume,
				Unit:     recipeIngredient.Unit,
				Meta:     recipeIngredient.Meta,
			})
		}
		dto.Parts = append(dto.Parts, dtoPart)
	}
	return dto
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

func (r *RecipeRepository) GetRecipes() ([]*Recipe, error) {
	db, err := db.Connect(r.cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var recipeDtos []*recipeDto
	err = db.Select(&recipeDtos, "SELECT * FROM recipes")
	if err != nil {
		return nil, err
	}
	recipes := make([]*Recipe, 0)
	for _, dto := range recipeDtos {
		recipes = append(recipes, mapRecipeDto(dto))
	}
	return recipes, nil
}

func (r *RecipeRepository) getRecipe(where string, args ...interface{}) (*Recipe, error) {
	db, err := db.Connect(r.cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var recipeDto recipeDto
	err = db.Get(&recipeDto, "SELECT * FROM recipes "+where, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return mapRecipeDto(&recipeDto), nil
}

func (r *RecipeRepository) GetRecipe(id string) (*Recipe, error) {
	uuidId, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id '%v': %w", id, err)
	}
	return r.getRecipe("WHERE id = $1", uuidId)
}

func (r *RecipeRepository) GetRecipeByTitle(title string) (*Recipe, error) {
	return r.getRecipe("WHERE title = $1", title)
}

func (r *RecipeRepository) SaveRecipe(recipe *Recipe) (*Recipe, error) {
	recipeUuid, err := uuid.Parse(recipe.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id '%v': %w", recipe.ID, err)
	}
	userUuid, err := uuid.Parse(recipe.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id '%v': %w", recipe.UserID, err)
	}
	db, err := db.Connect(r.cfg)
	if err != nil {
		return nil, err
	}
	dto := mapRecipe(recipe)
	defer db.Close()
	_, err = db.Exec("INSERT INTO recipes (id, title, description, image_id, user_id, published, tips, yield, parts)"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", recipeUuid, dto.Title, dto.Description, dto.ImageId, userUuid, dto.Published, pq.Array(dto.Tips), dto.Yield, dto.Parts)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	return r.GetRecipe(recipe.ID)
}
