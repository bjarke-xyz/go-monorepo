package recipes

import (
	"context"
	"fmt"
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
)

type RecipeService struct {
	recipeRepository model.RecipeRepository
	cache            *db.RedisCache
}

func NewRecipeService(recipeRepository model.RecipeRepository, cache *db.RedisCache) *RecipeService {
	return &RecipeService{
		recipeRepository: recipeRepository,
		cache:            cache,
	}
}

func (r *RecipeService) GetRecipes(ctx context.Context) ([]*model.Recipe, error) {
	cacheKey := "GetRecipes"
	var recipes []*model.Recipe
	if err := r.cache.Get(ctx, cacheKey, &recipes); err == nil {
		return recipes, nil
	}
	recipes, err := r.recipeRepository.GetRecipes(ctx)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, cacheKey, recipes, time.Hour)
	return recipes, nil
}

func (r *RecipeService) GetRecipeByTitle(ctx context.Context, title string) (*model.Recipe, error) {
	cacheKey := fmt.Sprintf("GetRecipeByTitle:%v", title)
	var recipe *model.Recipe
	if err := r.cache.Get(ctx, cacheKey, &recipe); err == nil {
		return recipe, nil
	}
	recipe, err := r.recipeRepository.GetRecipeByTitle(ctx, title)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, cacheKey, recipe, time.Hour)
	return recipe, nil
}

func (r *RecipeService) GetRecipe(ctx context.Context, id string) (*model.Recipe, error) {
	cacheKey := fmt.Sprintf("GetRecipe:%v", id)
	var recipe *model.Recipe
	if err := r.cache.Get(ctx, cacheKey, &recipe); err == nil {
		return recipe, nil
	}
	recipe, err := r.recipeRepository.GetRecipe(ctx, id)
	if err != nil {
		return nil, err
	}
	r.cache.Set(ctx, cacheKey, recipe, time.Hour)
	return recipe, nil
}

func (r *RecipeService) SaveRecipe(ctx context.Context, recipe *model.Recipe) (*model.Recipe, error) {
	return r.recipeRepository.SaveRecipe(ctx, recipe)
}
