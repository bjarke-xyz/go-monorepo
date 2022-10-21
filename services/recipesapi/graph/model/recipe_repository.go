package model

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type firebaseRecipeIngredientsDto struct {
	Original string   `firestore:"original,omitempty"`
	ID       *int     `firestore:"id,omitempty"`
	Title    *string  `firestore:"title,omitempty"`
	Volume   *float64 `firestore:"volume,omitempty"`
	Unit     *string  `firestore:"unit,omitempty"`
	Meta     []string `firestore:"meta,omitempty"`
}

type firebaseRecipeDto struct {
	ID             string                    `firestore:"id,omitempty"`
	Title          string                    `firestore:"title,omitempty"`
	Description    *string                   `firestore:"description,omitempty"`
	ImageId        *uuid.UUID                `firestore:"imageId,omitempty"`
	UserID         string                    `firestore:"userId,omitempty"`
	CreatedAt      time.Time                 `firestore:"createdAt,omitempty"`
	ModeratedAt    *time.Time                `firestore:"moderatedAt,omitempty"`
	LastModifiedAt time.Time                 `firestore:"lastModifiedAt,omitempty"`
	Published      bool                      `firestore:"published,omitempty"`
	Tips           []string                  `firestore:"tips,omitempty"`
	Yield          *string                   `firestore:"yield,omitempty"`
	Parts          []*firebaseRecipePartsDto `firestore:"parts,omitempty"`
}
type firebaseRecipePartsDto struct {
	Title       string                          `firestore:"title,omitempty"`
	Ingredients []*firebaseRecipeIngredientsDto `firestore:"ingredients,omitempty"`
	Steps       []string                        `firestore:"steps,omitempty"`
}

func mapFirebaseRecipeDto(dto firebaseRecipeDto) *Recipe {
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
func mapFirebaseRecipe(recipe *Recipe) *firebaseRecipeDto {
	dto := &firebaseRecipeDto{
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
		Parts:          make([]*firebaseRecipePartsDto, 0),
	}
	if dto.CreatedAt.IsZero() {
		dto.CreatedAt = time.Now()
	}
	if dto.LastModifiedAt.IsZero() {
		dto.LastModifiedAt = time.Now()
	}
	for _, recipePart := range recipe.Parts {
		recipePart := recipePart
		dtoPart := &firebaseRecipePartsDto{
			Title:       recipePart.Title,
			Steps:       recipePart.Steps,
			Ingredients: make([]*firebaseRecipeIngredientsDto, 0),
		}
		for _, recipeIngredient := range recipePart.Ingredients {
			recipeIngredient := recipeIngredient
			dtoPart.Ingredients = append(dtoPart.Ingredients, &firebaseRecipeIngredientsDto{
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

type NoSqlRecipeRepository struct {
	cfg    *config.Config
	app    *firebase.App
	client *firestore.Client
}

func NewNoSqlRecipeRepository(cfg *config.Config) RecipeRepository {
	noSqlRepo := &NoSqlRecipeRepository{
		cfg: cfg,
	}
	return noSqlRepo
}

const recipeCollection = "recipes"

func (r *NoSqlRecipeRepository) getClient(ctx context.Context) (*firestore.Client, error) {
	if r.app == nil {
		app, err := firebase.NewApp(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get firebase app: %w", err)
		}
		r.app = app
	}
	if r.app == nil {
		return nil, fmt.Errorf("firebase app was nil")
	}

	if r.client == nil {
		client, err := r.app.Firestore(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get firestore client: %w", err)
		}
		r.client = client
	}
	if r.client == nil {
		return nil, fmt.Errorf("firestore client was nil")
	}

	return r.client, nil
}

func (r *NoSqlRecipeRepository) GetRecipes(ctx context.Context) ([]*Recipe, error) {
	client, err := r.getClient(ctx)
	if err != nil {
		return nil, err
	}
	iter := client.Collection(recipeCollection).OrderBy("createdAt", firestore.Asc).Documents(ctx)
	recipes := make([]*Recipe, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get recipes: %w", err)
		}
		var dto firebaseRecipeDto
		err = doc.DataTo(&dto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse doc: %w", err)
		}
		recipe := mapFirebaseRecipeDto(dto)
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (r *NoSqlRecipeRepository) GetRecipeByTitle(ctx context.Context, title string) (*Recipe, error) {
	client, err := r.getClient(ctx)
	if err != nil {
		return nil, err
	}
	var dto *firebaseRecipeDto
	iter := client.Collection(recipeCollection).Where("title", "==", title).Limit(1).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get recipe: %w", err)
		}
		err = doc.DataTo(&dto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse doc: %w", err)
		}
	}
	if dto == nil {
		return nil, nil
	}
	recipe := mapFirebaseRecipeDto(*dto)
	return recipe, nil
}

func (r *NoSqlRecipeRepository) GetRecipesByUserId(ctx context.Context, userId string) ([]*Recipe, error) {
	client, err := r.getClient(ctx)
	if err != nil {
		return nil, err
	}
	iter := client.Collection(recipeCollection).Where("userId", "==", userId).Documents(ctx)
	recipes := make([]*Recipe, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get recipes: %w", err)
		}
		var dto firebaseRecipeDto
		err = doc.DataTo(&dto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse doc: %w", err)
		}
		recipe := mapFirebaseRecipeDto(dto)
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (r *NoSqlRecipeRepository) GetRecipe(ctx context.Context, id string) (*Recipe, error) {
	client, err := r.getClient(ctx)
	if err != nil {
		return nil, err
	}
	var dto firebaseRecipeDto
	doc, err := client.Collection(recipeCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed to get recipe with id '%v': %w", id, err)
		}
	}
	err = doc.DataTo(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc: %w", err)
	}
	recipe := mapFirebaseRecipeDto(dto)
	return recipe, nil
}

func (r *NoSqlRecipeRepository) SaveRecipe(ctx context.Context, recipe *Recipe) (*Recipe, error) {
	client, err := r.getClient(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapFirebaseRecipe(recipe)
	_, err = client.Collection(recipeCollection).Doc(recipe.ID).Set(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to add to recipe collection: %w", err)
	}
	return r.GetRecipe(ctx, recipe.ID)
}
