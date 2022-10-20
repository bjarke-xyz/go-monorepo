package file

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileDto struct {
	ID          uuid.UUID `db:"id" firestore:"id,omitempty"`
	Bucket      string    `db:"bucket" firestore:"bucket,omitempty"`
	Key         string    `db:"key" firestore:"key,omitempty"`
	ContentType string    `db:"content_type" firestore:"contentType,omitempty"`
	Size        int64     `db:"size" firestore:"size,omitempty"`
	Name        string    `db:"name" firestore:"name,omitempty"`
}

type FileRepository struct {
	cfg    *config.Config
	app    *firebase.App
	client *firestore.Client
}

const MaxFileSize = 5242880

func NewFileRepository(cfg *config.Config) *FileRepository {
	return &FileRepository{
		cfg: cfg,
	}
}

const fileCollection = "files"

func (f *FileRepository) getClient(ctx context.Context) (*firestore.Client, error) {
	if f.app == nil {
		app, err := firebase.NewApp(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get firebase app: %w", err)
		}
		f.app = app
	}
	if f.app == nil {
		return nil, fmt.Errorf("firebase app was nil")
	}

	if f.client == nil {
		client, err := f.app.Firestore(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get firestore client: %w", err)
		}
		f.client = client
	}
	if f.client == nil {
		return nil, fmt.Errorf("firestore client was nil")
	}

	return f.client, nil
}

func (f *FileRepository) GetFileById(ctx context.Context, id uuid.UUID) (*FileDto, error) {
	client, err := f.getClient(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := client.Collection(fileCollection).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed to get file: %w", err)
		}
	}
	var dto FileDto
	err = doc.DataTo(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc: %w", err)
	}
	return &dto, nil
}

func (f *FileRepository) GetFileByKey(ctx context.Context, bucket string, key string) (*FileDto, error) {
	client, err := f.getClient(ctx)
	if err != nil {
		return nil, err
	}
	var dto *FileDto
	iter := client.Collection(fileCollection).Where("bucket", "==", bucket).Where("key", "==", key).Limit(1).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get file: %w", err)
		}
		err = doc.DataTo(&dto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse doc: %w", err)
		}
	}
	return dto, nil
}

func (f *FileRepository) SaveFile(ctx context.Context, file *FileDto) error {
	client, err := f.getClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.Collection(fileCollection).Doc(file.ID.String()).Set(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
