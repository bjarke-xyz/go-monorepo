package list

import (
	"ShoppingList-Backend/internal/pkg/item"
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID        uuid.UUID  `db:"id" json:"id" validate:"required,uuid"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"`
	OwnerID   string     `db:"owner_id" json:"ownerId"`

	Name  string     `db:"name" json:"name"`
	Items []ListItem `db:"list_item" json:"items"`
}
type AddList struct {
	Name string `json:"name"`
}

type ListItem struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
	ListID    uuid.UUID  `db:"list_id" json:"listId"`
	ItemID    uuid.UUID  `db:"item_id" json:"itemId"`
	Item      item.Item  `db:"item" json:"item"`
	Crossed   bool       `db:"crossed" json:"crossed"`
}
type UpdateListItem struct {
	Crossed bool `json:"crossed"`
}

type DefaultList struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
	UserID    string     `db:"app_user_id" json:"userId"`
	ListID    uuid.UUID  `db:"list_id" json:"listId"`
}
