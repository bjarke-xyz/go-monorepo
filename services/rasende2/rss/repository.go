package rss

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/rasende2/pkg"
)

type RssRepository struct {
	context *pkg.AppContext
}

func NewRssRepository(context *pkg.AppContext) *RssRepository {
	return &RssRepository{
		context: context,
	}
}

type RssUrlDto struct {
	Name string   `json:"name"`
	Urls []string `json:"urls"`
}

func (r *RssRepository) GetRssUrls() ([]RssUrlDto, error) {
	jsonBytes, err := os.ReadFile("rss.json")
	if err != nil {
		return nil, fmt.Errorf("could not load rss.json: %w", err)
	}
	var rssUrls []RssUrlDto
	err = json.Unmarshal(jsonBytes, &rssUrls)
	if err != nil {
		return nil, err
	}
	return rssUrls, nil
}

type RssItemDto struct {
	ItemId    string    `db:"item_id" json:"itemId"`
	SiteName  string    `db:"site_name" json:"siteName"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Link      string    `db:"link" json:"link"`
	Published time.Time `db:"published" json:"published"`
}

func (r *RssRepository) SearchItems(query string, searchContent bool) ([]RssItemDto, error) {
	db, err := db.Connect(r.context.Config)
	if err != nil {
		return nil, err
	}
	db = db.Unsafe()
	defer db.Close()
	var rssItems []RssItemDto
	sql := "SELECT * FROM rss_items WHERE ts_title @@ to_tsquery('danish', $1)"
	if searchContent {
		sql = sql + " OR ts_content @@ to_tsquery('danish', $1)"
	}
	sql = sql + " ORDER BY published DESC"
	// err = db.Select(&rssItems, "SELECT * FROM rss_items WHERE LOWER(title) LIKE '%' || $1 || '%' order by published desc", query)
	err = db.Select(&rssItems, sql, query)
	if err != nil {
		return nil, fmt.Errorf("error getting items with query %v: %w", query, err)
	}
	return rssItems, nil
}

func (r *RssRepository) GetItems(siteName string) ([]RssItemDto, error) {
	db, err := db.Connect(r.context.Config)
	if err != nil {
		return nil, err
	}
	db = db.Unsafe()
	defer db.Close()
	var rssItems []RssItemDto
	err = db.Select(&rssItems, "SELECT * FROM rss_items WHERE site_name = $1", siteName)
	if err != nil {
		return nil, fmt.Errorf("error getting items for site %v: %w", siteName, err)
	}
	return rssItems, nil
}

func (r *RssRepository) InsertItems(items []RssItemDto) error {
	if len(items) == 0 {
		return nil
	}
	db, err := db.Connect(r.context.Config)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	_, err = db.NamedExec("INSERT INTO rss_items (item_id, site_name, title, content, link, published) "+
		"values (:item_id, :site_name, :title, :content, :link, :published) on conflict do nothing", items)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil

}
