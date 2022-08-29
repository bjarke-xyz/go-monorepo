package application

import (
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/internal/pkg/list"
)

type Controllers struct {
	Item *item.ItemController
	List *list.ListController
}
