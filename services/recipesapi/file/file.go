package file

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/google/uuid"
)

type FileDto struct {
	ID          uuid.UUID `db:"id"`
	Bucket      string    `db:"bucket"`
	Key         string    `db:"key"`
	ContentType string    `db:"content_type"`
	Size        int64     `db:"size"`
	Name        string    `db:"name"`
}

type FileRepository struct {
	cfg *config.Config
}

func NewFileRepository(cfg *config.Config) *FileRepository {
	return &FileRepository{
		cfg: cfg,
	}
}

func (f *FileRepository) getFileWhere(where string, args ...interface{}) (*FileDto, error) {
	db, err := db.Connect(f.cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var fileDto FileDto
	err = db.Get(&fileDto, "SELECT * FROM files "+where, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &fileDto, nil
}

func (f *FileRepository) GetFileById(id uuid.UUID) (*FileDto, error) {
	return f.getFileWhere("WHERE id = $1", id)
}

func (f *FileRepository) GetFileByKey(bucket string, key string) (*FileDto, error) {
	return f.getFileWhere("WHERE bucket = $1 AND key = $2", bucket, key)
}

func (f *FileRepository) SaveFile(file *FileDto) error {
	db, err := db.Connect(f.cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.NamedExec("INSERT INTO files (id, bucket, key, content_type, size, name) VALUES (:id, :bucket, :key, :content_type, :size, :name)", file)
	if err != nil {
		return fmt.Errorf("failed to insert file: %w", err)
	}
	return nil
}
