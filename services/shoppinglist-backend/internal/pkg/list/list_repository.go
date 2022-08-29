package list

import (
	"ShoppingList-Backend/internal/pkg/item"
	"ShoppingList-Backend/internal/pkg/user"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ListRepository struct {
	DB *sqlx.DB
}

func (q *ListRepository) getListItems(listIds []uuid.UUID) ([]ListItem, error) {
	listItems := []ListItem{}
	query, args, err := sqlx.In(`SELECT * FROM list_item WHERE list_id IN (?) ORDER BY updated_at desc`, listIds)
	if err != nil {
		return listItems, err
	}

	query = q.DB.Rebind(query)
	err = q.DB.Select(&listItems, query, args...)
	if err != nil {
		return listItems, err
	}

	return listItems, nil
}

func (q *ListRepository) getItems(itemIds []uuid.UUID) ([]item.Item, error) {
	items := []item.Item{}
	query, args, err := sqlx.In(`SELECT * FROM items WHERE id IN (?)`, itemIds)
	if err != nil {
		return items, err
	}

	query = q.DB.Rebind(query)
	err = q.DB.Select(&items, query, args...)
	if err != nil {
		return items, err
	}
	return items, nil
}

func (q *ListRepository) populateWithItems(lists []List) error {
	// Get ListItems
	listIds := make([]uuid.UUID, len(lists))
	for _, list := range lists {
		listIds = append(listIds, list.ID)
	}
	listItems, err := q.getListItems(listIds)
	if err != nil {
		return err
	}
	listItemsByListId := make(map[uuid.UUID][]ListItem)
	for _, listItem := range listItems {
		_, present := listItemsByListId[listItem.ListID]
		if present {
			listItemsByListId[listItem.ListID] = append(listItemsByListId[listItem.ListID], listItem)
		} else {
			listItemsByListId[listItem.ListID] = []ListItem{listItem}
		}
	}
	for i, list := range lists {
		lists[i].Items = listItemsByListId[list.ID]
	}

	// Get ListItem.Item
	itemIds := make([]uuid.UUID, len(lists))
	for _, list := range lists {
		for _, listItem := range list.Items {
			itemIds = append(itemIds, listItem.ItemID)
		}
	}
	items, err := q.getItems(itemIds)
	if err != nil {
		return err
	}
	itemsById := make(map[uuid.UUID]item.Item)
	for _, item := range items {
		itemsById[item.ID] = item
	}
	for i := range lists {
		for j, item := range lists[i].Items {
			lists[i].Items[j].Item = itemsById[item.ItemID]
		}
	}

	return nil
}

func (q *ListRepository) GetLists(owner *user.AppUser) ([]List, error) {
	lists := []List{}

	query := `SELECT * FROM lists WHERE owner_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC`

	err := q.DB.Select(&lists, query, owner.ID)

	if err != nil {
		return lists, err
	}

	q.populateWithItems(lists)

	for i := range lists {
		if lists[i].Items == nil {
			lists[i].Items = make([]ListItem, 0)
		}
	}

	return lists, nil
}

func (q *ListRepository) DeleteLists(ownerID string) error {
	query := `DELETE FROM LISTS where owner_id = $1`
	_, err := q.DB.Exec(query, ownerID)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) GetList(id uuid.UUID, appUser *user.AppUser) (List, error) {
	list := List{}

	query := `SELECT * FROM lists WHERE id = $1 AND deleted_at IS NULL LIMIT 1`

	err := q.DB.Get(&list, query, id)
	if err != nil {
		return list, err
	}

	// TODO: check if list is shared with user
	if list.OwnerID != appUser.ID /* || list is shared with user */ {
		return list, errors.New("access not allowed")
	}

	lists := []List{list}
	q.populateWithItems(lists)

	if lists[0].Items == nil {
		lists[0].Items = make([]ListItem, 0)
	}

	return lists[0], err
}

func (q *ListRepository) CreateList(list List) (uuid.UUID, error) {
	query := `INSERT INTO lists (id, name, owner_id) VALUES ($1, $2, $3)`
	_, err := q.DB.Exec(query, list.ID, list.Name, list.OwnerID)
	if err != nil {
		return uuid.Nil, err
	}

	return list.ID, nil
}

func (q *ListRepository) UpdateList(list List) error {
	query := `UPDATE lists SET updated_at = NOW(), name = $2 WHERE id = $1`
	_, err := q.DB.Exec(query, list.ID, list.Name)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) AddItemToList(list List, item item.Item) (ListItem, error) {
	listItem := ListItem{ID: uuid.New()}
	query := `INSERT INTO list_item (id, list_id, item_id) VALUES ($1, $2, $3)`
	_, err := q.DB.Exec(query, listItem.ID, list.ID, item.ID)
	if err != nil {
		return listItem, err
	}
	fetchQuery := `SELECT * FROM list_item WHERE id = $1`
	err = q.DB.Get(&listItem, fetchQuery, listItem.ID)
	if err != nil {
		return listItem, err
	}
	listItem.Item = item
	return listItem, nil
}

func (q *ListRepository) UpdateListItem(listItem ListItem) error {
	query := `UPDATE list_item SET updated_at = NOW(), crossed = $1 WHERE id = $2`
	_, err := q.DB.Exec(query, listItem.Crossed, listItem.ID)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) GetListItem(id uuid.UUID) (ListItem, error) {
	listItem := ListItem{}
	query := `SELECT * FROM list_item where id = $1`
	err := q.DB.Get(&listItem, query, id)
	if err != nil {
		return listItem, err
	}
	item := item.Item{}
	itemQuery := `SELECT * FROM items WHERE id = $1`
	err = q.DB.Get(&item, itemQuery, listItem.ItemID)
	if err != nil {
		return listItem, err
	}
	listItem.Item = item
	return listItem, nil
}

func (q *ListRepository) RemoveItemFromList(id uuid.UUID) error {
	query := `DELETE FROM list_item WHERE id = $1`
	_, err := q.DB.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) DeleteList(list List) error {
	query := `UPDATE lists SET deleted_at = NOW() WHERE id = $1`
	_, err := q.DB.Exec(query, list.ID)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) DeleteCrossedListItems(list List) error {
	query := `DELETE FROM list_item WHERE list_id = $1 AND crossed = true`
	_, err := q.DB.Exec(query, list.ID)
	if err != nil {
		return err
	}
	return nil
}

func (q *ListRepository) GetDefaultList(user *user.AppUser) (DefaultList, error) {
	fetchQuery := `SELECT * FROM default_lists WHERE app_user_id = $1 LIMIT 1`
	defaultList := DefaultList{}
	err := q.DB.Get(&defaultList, fetchQuery, user.ID)
	if err != nil {
		return defaultList, err
	}
	return defaultList, nil
}

func (q *ListRepository) ClearDefaultList(user *user.AppUser) error {
	query := `DELETE FROM default_lists WHERE app_user_id = $1`
	_, err := q.DB.Exec(query, user.ID)
	return err
}

func (q *ListRepository) SetDefaultList(user *user.AppUser, list List) (DefaultList, error) {
	fetchQuery := `SELECT * FROM default_lists WHERE app_user_id = $1 LIMIT 1`
	currentDefaultList := DefaultList{}
	err := q.DB.Get(&currentDefaultList, fetchQuery, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User does not have a default list, create one
			insertQuery := `INSERT INTO default_lists (app_user_id, list_id) VALUES ($1, $2)`
			_, err := q.DB.Exec(insertQuery, user.ID, list.ID)
			if err != nil {
				return currentDefaultList, err
			}
		} else {
			return currentDefaultList, err
		}
	} else if currentDefaultList.ListID != list.ID {
		// User already has a default list, so update it
		updateQuery := `UPDATE default_lists SET list_id = $2, updated_at = NOW() WHERE app_user_id = $1`
		_, err := q.DB.Exec(updateQuery, user.ID, list.ID)
		if err != nil {
			return currentDefaultList, err
		}
	}

	// If the default list was updated or created, fetch again
	if currentDefaultList.ListID != list.ID {
		err = q.DB.Get(&currentDefaultList, fetchQuery, user.ID)
		if err != nil {
			return currentDefaultList, err
		}
	}
	return currentDefaultList, nil
}
