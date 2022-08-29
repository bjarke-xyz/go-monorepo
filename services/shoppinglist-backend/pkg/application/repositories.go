package application

import (
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/internal/pkg/list"
)

type Repositories struct {
	Item *item.ItemRepository
	List *list.ListRepository
}
