package items

import (
	"ShoppingList-Backend/internal/pkg/common"
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/pkg/application"
	"ShoppingList-Backend/pkg/middleware"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetItems func gets all items for user
// @Description Get all items for user
// @Summary get all items for user
// @Tags items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=[]item.Item}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Router /api/v1/items [get]
func GetItems(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appUser := middleware.UserFromContext(r.Context())

		items, err := app.Controllers.Item.GetItems(appUser)
		if err != nil {
			app.Srv.RespondError(w, r, err.StatusCode, err.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: items,
		})
	}
}

// CreateItem func Create new item
// @Description Create new item
// @Summary Create new item
// @Tags items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param item body item.AddItem true "Add item"
// @Success 200 {object} common.Response{data=item.Item}
// @Failure 500 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/items [post]
func CreateItem(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		addItem := &item.AddItem{}
		if err := app.Srv.Decode(w, r, addItem); err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err))
			return
		}
		appUser := middleware.UserFromContext(r.Context())

		createdItem, err := app.Controllers.Item.CreateItem(appUser, addItem)
		if err != nil {
			app.Srv.RespondError(w, r, err.StatusCode, err.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: createdItem,
		})
	}
}

// UpdateItem func Update item
// @Description Update item
// @Summary Update item
// @Tags items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param item body item.AddItem true "Update item"
// @Success 200 {object} common.Response{data=item.Item}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/items/{id} [put]
func UpdateItem(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse item id %v: %w", idStr, err))
			return
		}

		appUser := middleware.UserFromContext(r.Context())

		addItem := &item.AddItem{}
		if err := app.Srv.Decode(w, r, addItem); err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err))
			return
		}

		updatedItem, errr := app.Controllers.Item.UpdateItem(appUser, id, addItem)
		if err != nil {
			app.Srv.RespondError(w, r, errr.StatusCode, errr.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: updatedItem,
		})
	}
}

// DeleteItem func Delete item
// @Description Delete item
// @Summary Delete item
// @Tags items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Success 204 {string} status "ok"
// @Failure 500 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/items/{id} [delete]
func DeleteItem(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse item id %v: %w", idStr, err))
			return
		}

		appUser := middleware.UserFromContext(r.Context())

		_, cErr := app.Controllers.Item.DeleteItem(appUser, id)
		if cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusNoContent, nil)
	}

}
