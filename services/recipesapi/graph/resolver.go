//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"time"

	"github.com/bjarke-xyz/go-monorepo/services/recipesapi/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	recipeRepository *model.RecipeRepository
	userRepository   *model.UserRepository
}

func StringPtr(str string) *string {
	return &str
}
func Float64Ptr(val int) *float64 {
	floatVal := float64(val)
	return &floatVal
}

func NewResolver(userRepository *model.UserRepository, recipeRepository *model.RecipeRepository) *Resolver {
	resolver := &Resolver{
		recipeRepository: recipeRepository,
		userRepository:   userRepository,
	}
	InsertFakeData(resolver)
	return resolver
}

func InsertFakeData(resolver *Resolver) {
	user, _ := resolver.userRepository.CreateUser("Kokkehuen")
	recipe := &model.Recipe{
		Title:          "Tiramusi",
		Description:    StringPtr("Tiramisu er en klassisk norditaliensk dessert. Desserten er opbygget i lag bestående af ladyfingers, som er vædet med stærk kaffe, og mascarponecreme. Dette er den tidligst kendte opskrift på tiramisu, som stammer fra restaurant Le Beccherie. Opskriften blev offentliggjort i et nummer af magasinet Vin Veneto trykt i foråret 1981. "),
		Image:          nil,
		UserID:         user.ID,
		CreatedAt:      time.Now(),
		ModeratedAt:    nil,
		LastModifiedAt: time.Now(),
		Published:      true,
		Tips:           []string{"Tilsæt Marsala eller Amaretto til kaffen. "},
		Yield:          nil,
		Parts: []*model.RecipeParts{
			&model.RecipeParts{
				Title: "Tiramisu",
				Ingredients: []*model.RecipeIngredients{
					&model.RecipeIngredients{
						Title:  StringPtr("Æggeblommer"),
						Volume: Float64Ptr(6),
					},
					&model.RecipeIngredients{
						Title:  StringPtr("Sukker"),
						Unit:   StringPtr("gram"),
						Volume: Float64Ptr(250),
					},
					&model.RecipeIngredients{
						Title:  StringPtr("Mascarpone"),
						Unit:   StringPtr("gram"),
						Volume: Float64Ptr(500),
					},
					&model.RecipeIngredients{
						Title:  StringPtr("Ladyfingers"),
						Unit:   StringPtr("gram"),
						Volume: Float64Ptr(200),
					},
					&model.RecipeIngredients{
						Title:  StringPtr("Stærk kaffe"),
						Unit:   StringPtr("deciliter"),
						Volume: Float64Ptr(4),
					},
				},
				Steps: []string{
					"Pisk æggeblommer med sukker til en lys og luftig æggesnaps. ",
					"Tilsæt mascarpone og pisk til en blød creme. ",
					"Dyp halvdelen af dine ladyfingers i kaffen, og placer dem i ét lag i bunden af et fad. De skal dække hele bunden, og ligge så tæt som muligt. Det kan være nødvendigt at knække dem i mindre stykker. ",
					"Fordel halvdelen af cremen over dine ladyfingers. ",
					"Dyp de resterende ladyfingers i kaffen, og placer dem i endnu et lag. Det er en god idé et lægge dem på tværs, da dette vil være med til at holde på kagen. ",
					"Fordel resten af cremen over dine ladyfingers.",
					"Sæt kagen på køl natten over. ",
				},
			},
			&model.RecipeParts{
				Title: "Pynt",
				Ingredients: []*model.RecipeIngredients{
					&model.RecipeIngredients{
						Title:  StringPtr("Kakaopulver"),
						Unit:   StringPtr("spiseske"),
						Volume: Float64Ptr(2),
					},
				},
				Steps: []string{
					"Drys med kakaopulver lige inden servering. ",
				},
			},
		},
	}
	resolver.recipeRepository.CreateRecipe(recipe)
}
