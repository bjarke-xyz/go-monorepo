package list

import (
	"ShoppingList-Backend/internal/pkg/controller"
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/internal/pkg/user"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type ListController struct {
	itemRepo *item.ItemRepository
	listRepo *ListRepository
}

func NewListController(itemRepo *item.ItemRepository, listRepo *ListRepository) *ListController {
	return &ListController{
		itemRepo: itemRepo,
		listRepo: listRepo,
	}
}

func (c *ListController) GetLists(user *user.AppUser) ([]List, *controller.ControllerError) {
	lists, err := c.listRepo.GetLists(user)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("lists not found: %w", err))
	}

	return lists, nil
}

func (c *ListController) GetDefaultList(user *user.AppUser) (*DefaultList, *controller.ControllerError) {
	defaultList, err := c.listRepo.GetDefaultList(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, controller.CError(http.StatusNotFound, fmt.Errorf("no default list"))
		} else {
			return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("error getting default list for user ID %v: %w", user.ID, err))
		}
	}

	return &defaultList, nil
}

func (c *ListController) CreateList(user *user.AppUser, addList *AddList) (*List, *controller.ControllerError) {
	listToCreate := List{
		ID:      uuid.New(),
		Name:    addList.Name,
		OwnerID: user.ID,
	}

	listId, err := c.listRepo.CreateList(listToCreate)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not create list: %w", err))
	}

	createdList, err := c.listRepo.GetList(listId, user)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not get created list with ID %v: %w", listId, err))
	}

	return &createdList, nil
}

func (c *ListController) UpdateList(user *user.AppUser, listID uuid.UUID, updateList *AddList) (*List, *controller.ControllerError) {
	foundList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	foundList.Name = updateList.Name
	if err := c.listRepo.UpdateList(foundList); err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not update list with ID %v: %w", listID, err))
	}

	updatedList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not get updated list with ID %v: %w", listID, err))
	}

	return &updatedList, nil
}

func (c *ListController) SetDefaultList(user *user.AppUser, listID uuid.UUID) (*DefaultList, *controller.ControllerError) {
	foundList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	defaultList, err := c.listRepo.SetDefaultList(user, foundList)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not set default list with ID %v: %w", listID, err))
	}

	return &defaultList, nil
}

func (c *ListController) AddItemToList(user *user.AppUser, listID uuid.UUID, itemID uuid.UUID) (*ListItem, *controller.ControllerError) {
	foundList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	foundItem, err := c.itemRepo.GetItem(itemID)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("item with ID %v not found: %w", listID, err))
	}

	listItem, err := c.listRepo.AddItemToList(foundList, foundItem)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not add item (%v) to list (%v): %w", itemID, listID, err))
	}

	return &listItem, nil
}

func (c *ListController) UpdateListItem(user *user.AppUser, listID uuid.UUID, listItemID uuid.UUID, updateListItem *UpdateListItem) (*ListItem, *controller.ControllerError) {
	if _, err := c.listRepo.GetList(listID, user); err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	listItem, err := c.listRepo.GetListItem(listItemID)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("listItem with ID %v not found: %w", listItemID, err))
	}

	listItem.Crossed = updateListItem.Crossed
	if err := c.listRepo.UpdateListItem(listItem); err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not update ListItem with ID %v: %w", listItemID, err))
	}

	return &listItem, nil
}

func (c *ListController) RemoveItemFromList(user *user.AppUser, listID uuid.UUID, listItemID uuid.UUID) *controller.ControllerError {
	if _, err := c.listRepo.GetList(listID, user); err != nil {
		return controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	if err := c.listRepo.RemoveItemFromList(listItemID); err != nil {
		return controller.CError(http.StatusInternalServerError, fmt.Errorf("could not remove listitem (%v) from list (%v): %w", listItemID, listID, err))
	}

	return nil
}

func (c *ListController) DeleteList(user *user.AppUser, listID uuid.UUID) *controller.ControllerError {
	foundList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	if err := c.listRepo.DeleteList(foundList); err != nil {
		return controller.CError(http.StatusInternalServerError, fmt.Errorf("could not delete list (%v): %w", listID, err))
	}

	defaultList, err := c.listRepo.GetDefaultList(user)
	if err != nil {
		return controller.CError(http.StatusInternalServerError, fmt.Errorf("could not get default list: %w", err))
	}

	if defaultList.ListID == foundList.ID {
		err = c.listRepo.ClearDefaultList(user)
		if err != nil {
			return controller.CError(http.StatusInternalServerError, fmt.Errorf("could not clear default list: %w", err))
		}
	}

	return nil
}

func (c *ListController) DeleteCrossedListItems(user *user.AppUser, listID uuid.UUID) *controller.ControllerError {
	foundList, err := c.listRepo.GetList(listID, user)
	if err != nil {
		return controller.CError(http.StatusNotFound, fmt.Errorf("list with ID %v not found: %w", listID, err))
	}

	if err := c.listRepo.DeleteCrossedListItems(foundList); err != nil {
		return controller.CError(http.StatusInternalServerError, fmt.Errorf("could not delete crossed list items (%v): %w", listID, err))
	}

	return nil
}
