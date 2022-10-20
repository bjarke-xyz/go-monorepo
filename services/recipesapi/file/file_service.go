package file

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bjarke-xyz/go-monorepo/libs/common/storage"
	"github.com/google/uuid"
)

type FileService struct {
	fileRepository *FileRepository
	storage        *storage.StorageClient
}

func NewFileService(fileRepository *FileRepository, storage *storage.StorageClient) *FileService {
	return &FileService{
		fileRepository: fileRepository,
		storage:        storage,
	}
}

func (f *FileService) SaveImage(ctx context.Context, recipeId string, image *graphql.Upload) (uuid.UUID, error) {
	if image == nil {
		return uuid.Nil, nil
	}
	if !strings.HasPrefix(image.ContentType, "image") {
		return uuid.Nil, fmt.Errorf("file must be an image")
	}
	if image.Size > MaxFileSize {
		return uuid.Nil, fmt.Errorf("image must be less than 5 megabytes")
	}
	imageData, err := io.ReadAll(image.File)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not read image file: %w", err)
	}
	imageId := uuid.New()
	bucket := "recipesapi"
	key := fmt.Sprintf("/images/%v/%v", recipeId, imageId.String())
	err = f.storage.Put(ctx, bucket, key, imageData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to store image: %w", err)
	}
	fileDto := &FileDto{
		ID:          imageId,
		Bucket:      bucket,
		Key:         key,
		ContentType: image.ContentType,
		Size:        image.Size,
		Name:        image.Filename,
	}
	err = f.fileRepository.SaveFile(ctx, fileDto)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save image info: %w", err)
	}
	return imageId, nil
}

func (f *FileService) GetFileById(ctx context.Context, id uuid.UUID) (*FileDto, error) {
	return f.fileRepository.GetFileById(ctx, id)
}
