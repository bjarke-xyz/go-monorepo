package item

import (
	"ShoppingList-Backend/internal/pkg/controller"
	"ShoppingList-Backend/internal/pkg/user"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type ItemController struct {
	itemRepo *ItemRepository
}

func NewItemController(itemRepo *ItemRepository) *ItemController {
	return &ItemController{
		itemRepo: itemRepo,
	}
}

func (c *ItemController) GetItems(user *user.AppUser) ([]Item, *controller.ControllerError) {
	if user == nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("nil user"))
	}
	items, err := c.itemRepo.GetItems(user.ID)
	if err != nil {
		return nil, controller.CError(http.StatusBadRequest, fmt.Errorf("items not found: %w", err))
	}
	return items, nil
}

func (c *ItemController) CreateItem(user *user.AppUser, addItem *AddItem) (*Item, *controller.ControllerError) {
	if user == nil || addItem == nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("nil params: %v and %v", user, addItem))
	}

	itemToCreate := &Item{
		ID:      uuid.New(),
		Name:    addItem.Name,
		OwnerID: user.ID,
	}

	itemId, err := c.itemRepo.CreateItem(itemToCreate)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not create item: %w", err))
	}

	createdItem, err := c.itemRepo.GetItem(itemId)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not get created item: %w", err))
	}

	return &createdItem, nil
}

func (c *ItemController) UpdateItem(user *user.AppUser, itemID uuid.UUID, updateItem *AddItem) (*Item, *controller.ControllerError) {
	if user == nil || updateItem == nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("nil params: %v and %v", user, updateItem))
	}

	foundItem, err := c.itemRepo.GetItem(itemID)
	if err != nil {
		return nil, controller.CError(http.StatusNotFound, fmt.Errorf("item with ID %v not found: %w", itemID, err))
	}

	foundItem.Name = updateItem.Name

	if err := c.itemRepo.UpdateItem(&foundItem); err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not update item ID %v: %w", itemID, err))
	}

	updatedItem, err := c.itemRepo.GetItem(itemID)
	if err != nil {
		return nil, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not get updated item with ID %v: %w", itemID, err))
	}

	return &updatedItem, nil
}

func (c *ItemController) DeleteItem(user *user.AppUser, itemID uuid.UUID) (bool, *controller.ControllerError) {

	foundItem, err := c.itemRepo.GetItem(itemID)
	if err != nil {
		return false, controller.CError(http.StatusNotFound, fmt.Errorf("item with ID %v not found: %w", itemID, err))
	}

	if foundItem.OwnerID != user.ID {
		return false, controller.CError(http.StatusNotFound, fmt.Errorf("item with ID %v not found", itemID))
	}

	if err := c.itemRepo.DeleteItem(&foundItem); err != nil {
		return false, controller.CError(http.StatusInternalServerError, fmt.Errorf("could not delete item with ID %v: %w", itemID, err))
	}

	return true, nil
}
