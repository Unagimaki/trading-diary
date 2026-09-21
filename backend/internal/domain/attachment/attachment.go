package attachment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const MaxSize int64 = 5 << 20

var ErrNotFound = errors.New("attachment target not found")
var ErrInvalidFile = errors.New("invalid image")

type Attachment struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mimeType"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}
type Stored struct {
	Attachment
	StorageKey string
}
type Storage interface {
	Save(context.Context, io.Reader) (string, error)
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
type Repository interface {
	Upsert(context.Context, string, string, string, string, string, string, string, int64) (Attachment, string, error)
	Get(context.Context, string, string) (Stored, error)
	Delete(context.Context, string, string, string, string) (string, error)
}
type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}
func (s *Service) Upload(ctx context.Context, userID, journalID, rowID, columnID, name string, size int64, source io.Reader) (Attachment, error) {
	if size <= 0 || size > MaxSize {
		return Attachment{}, ErrInvalidFile
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(source, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return Attachment{}, ErrInvalidFile
	}
	head = head[:n]
	mime := http.DetectContentType(head)
	if mime != "image/jpeg" && mime != "image/png" && mime != "image/webp" {
		return Attachment{}, ErrInvalidFile
	}
	key, err := s.storage.Save(ctx, io.MultiReader(bytes.NewReader(head), source))
	if err != nil {
		return Attachment{}, err
	}
	item, oldKey, err := s.repo.Upsert(ctx, userID, journalID, rowID, columnID, key, cleanName(name), mime, size)
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		return Attachment{}, err
	}
	if oldKey != "" && oldKey != key {
		_ = s.storage.Delete(ctx, oldKey)
	}
	item.MimeType = mime
	return item, nil
}
func (s *Service) Open(ctx context.Context, userID, id string) (Stored, io.ReadCloser, error) {
	item, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return item, nil, err
	}
	reader, err := s.storage.Open(ctx, item.StorageKey)
	return item, reader, err
}
func (s *Service) Delete(ctx context.Context, userID, journalID, rowID, columnID string) error {
	key, err := s.repo.Delete(ctx, userID, journalID, rowID, columnID)
	if err != nil {
		return err
	}
	if key != "" {
		return s.storage.Delete(ctx, key)
	}
	return nil
}
func cleanName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "" {
		return "image"
	}
	if len([]rune(name)) > 255 {
		return string([]rune(name)[:255])
	}
	return name
}
