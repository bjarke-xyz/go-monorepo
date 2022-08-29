package item

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID        uuid.UUID  `db:"id" json:"id" validate:"required,uuid"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"`
	OwnerID   string     `db:"owner_id" json:"ownerId"`

	Name string `db:"name" json:"name"`
}

type AddItem struct {
	Name string `json:"name"`
}
