package lists

import (
	"ShoppingList-Backend/internal/pkg/common"
	"ShoppingList-Backend/internal/pkg/list"
	"ShoppingList-Backend/pkg/application"
	"ShoppingList-Backend/pkg/middleware"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetLists func gets all lists for user
// @Description Get all lists for user
// @Summary get all lists for user
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=[]list.List}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Router /api/v1/lists [get]
func GetLists(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appUser := middleware.UserFromContext(r.Context())
		lists, err := app.Controllers.List.GetLists(appUser)
		if err != nil {
			app.Srv.RespondError(w, r, err.StatusCode, err.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: lists,
		})
	}
}

// GetDefaultList func Get the user's default list
// @Description Get the user's default list
// @Summary Get the user's default list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=list.DefaultList}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Router /api/v1/lists/default [get]
func GetDefaultList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appUser := middleware.UserFromContext(r.Context())
		defaultList, err := app.Controllers.List.GetDefaultList(appUser)
		if err != nil {
			app.Srv.RespondError(w, r, err.StatusCode, err.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: defaultList,
		})
	}
}

// CreateList func Create new list
// @Description Create new list
// @Summary Create new list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param list body list.AddList true "Add list"
// @Success 200 {object} common.Response{data=list.List}
// @Failure 500 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists [post]
func CreateList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addList := &list.AddList{}
		if err := app.Srv.Decode(w, r, addList); err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not decode body: %w", err))
			return
		}

		appUser := middleware.UserFromContext(r.Context())

		createdList, err := app.Controllers.List.CreateList(appUser, addList)
		if err != nil {
			app.Srv.RespondError(w, r, err.StatusCode, err.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: createdList,
		})
	}
}

// UpdateList func Update list
// @Description Update list
// @Summary Update list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param list body list.AddList true "Update list"
// @Success 200 {object} common.Response{data=list.List}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{id} [put]
func UpdateList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse id %v: %w", idStr, err))
			return
		}

		updateList := &list.AddList{}
		if err := app.Srv.Decode(w, r, updateList); err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err))
			return
		}

		appUser := middleware.UserFromContext(r.Context())

		updatedList, cErr := app.Controllers.List.UpdateList(appUser, id, updateList)
		if cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		// app.SseBroker.Notifier <- sse.NewNotification(
		// 	sse.CreateEvent(sse.BrokerEvent{
		// 		EventType:  list.EventListUpdated,
		// 		EventData:  updatedList,
		// 		Recipients: []string{appUser.ID},
		// 	}),
		// )

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: updatedList,
		})
	}
}

// SetDefaultList func set default list
// @Description set default list
// @Summary set default list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} common.Response{data=list.DefaultList}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{id}/default [put]
func SetDefaultList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse id %v: %w", idStr, err))
			return
		}

		appUser := middleware.UserFromContext(r.Context())
		defaultList, cErr := app.Controllers.List.SetDefaultList(appUser, id)
		if cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: defaultList,
		})
	}
}

// AddItemToList func Add item to list
// @Description Add item to list
// @Summary Add item to list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param list-id path string true "List ID"
// @Param item-id path string true "Item ID"
// @Success 200 {object} common.Response{data=list.ListItem}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{list-id}/items/{item-id} [post]
func AddItemToList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		listIdStr := params["id"]
		listId, err := uuid.Parse(listIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse list id %v: %w", listIdStr, err))
			return
		}

		itemIdStr := params["itemId"]
		itemId, err := uuid.Parse(itemIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse item id %v: %w", itemIdStr, err))
			return
		}

		user := middleware.UserFromContext(r.Context())

		listItem, cErr := app.Controllers.List.AddItemToList(user, listId, itemId)
		if cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		// app.SseBroker.Notifier <- sse.NewNotification(
		// 	sse.CreateEvent(sse.BrokerEvent{
		// 		EventType:  list.EventListItemsAdded,
		// 		EventData:  listItem,
		// 		Recipients: []string{user.ID},
		// 	}),
		// )

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: listItem,
		})
	}
}

// UpdateListItem func Update list item
// @Description Update list item
// @Summary Update list item
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param list-id path string true "List ID"
// @Param list-item-id path string true "List-Item ID"
// @Param list body list.UpdateListItem true "Update list item"
// @Success 200 {object} common.Response{data=list.ListItem}
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{list-id}/items/{list-item-id} [put]
func UpdateListItem(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		listIdStr := params["id"]
		listId, err := uuid.Parse(listIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse list id %v: %w", listIdStr, err))
			return
		}

		listItemIdStr := params["listItemId"]
		listItemId, err := uuid.Parse(listItemIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse listitem id %v: %w", listItemIdStr, err))
			return
		}

		updateListItem := &list.UpdateListItem{}
		if err := app.Srv.Decode(w, r, updateListItem); err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err))
			return
		}

		user := middleware.UserFromContext(r.Context())

		updatedListItem, cErr := app.Controllers.List.UpdateListItem(user, listId, listItemId, updateListItem)
		if cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		// app.SseBroker.Notifier <- sse.NewNotification(
		// 	sse.CreateEvent(sse.BrokerEvent{
		// 		EventType:  list.EventListItemsUpdated,
		// 		EventData:  updatedListItem,
		// 		Recipients: []string{user.ID},
		// 	}),
		// )

		app.Srv.Respond(w, r, http.StatusOK, common.Response{
			Data: updatedListItem,
		})
	}
}

// RemoveItemFromList func Remove item from list
// @Description Remove item from list
// @Summary Remove item from list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param list-id path string true "List ID"
// @Param list-item-id path string true "List-Item ID"
// @Success 204 {string} status "ok"
// @Failure 500 {object} server.HTTPError
// @Failure 404 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{list-id}/items/{list-item-id} [delete]
func RemoveItemFromList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		listIdStr := params["id"]
		listId, err := uuid.Parse(listIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse list id %v: %w", listIdStr, err))
			return
		}

		listItemIdStr := params["listItemId"]
		listItemId, err := uuid.Parse(listItemIdStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse listitem id %v: %w", listItemIdStr, err))
			return
		}

		user := middleware.UserFromContext(r.Context())

		if cErr := app.Controllers.List.RemoveItemFromList(user, listId, listItemId); cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		// app.SseBroker.Notifier <- sse.NewNotification(
		// 	sse.CreateEvent(sse.BrokerEvent{
		// 		EventType:  list.EventListItemsRemoved,
		// 		EventData:  listItemId,
		// 		Recipients: []string{user.ID},
		// 	}),
		// )

		app.Srv.Respond(w, r, http.StatusNoContent, nil)
	}
}

// DeleteList func Delete list
// @Description Delete list
// @Summary Delete list
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 204 {string} status "ok"
// @Failure 500 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{id} [delete]
func DeleteList(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse id %v: %w", idStr, err))
			return
		}

		user := middleware.UserFromContext(r.Context())

		if cErr := app.Controllers.List.DeleteList(user, id); cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusNoContent, nil)
	}
}

// ClearCrossedListItems func Clear crossed list items
// @Description Clear crossed list items
// @Summary Clear crossed list items
// @Tags lists
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 204 {string} status "ok"
// @Failure 500 {object} server.HTTPError
// @Failure 400 {object} server.HTTPError
// @Router /api/v1/lists/{id}/items/crossed [delete]
func ClearCrossedListItems(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.Srv.RespondError(w, r, http.StatusBadRequest, fmt.Errorf("could not parse id %v: %w", idStr, err))
			return
		}

		user := middleware.UserFromContext(r.Context())

		if cErr := app.Controllers.List.DeleteCrossedListItems(user, id); cErr != nil {
			app.Srv.RespondError(w, r, cErr.StatusCode, cErr.Err)
			return
		}

		app.Srv.Respond(w, r, http.StatusNoContent, nil)
	}
}
