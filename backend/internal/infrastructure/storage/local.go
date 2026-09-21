package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type Local struct{ root string }

func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	return &Local{root: root}, nil
}
func (s *Local) Save(_ context.Context, source io.Reader) (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	key := hex.EncodeToString(raw)
	path := filepath.Join(s.root, key)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return "", copyErr
	}
	return key, closeErr
}
func (s *Local) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.root, filepath.Base(key)))
}
func (s *Local) Delete(_ context.Context, key string) error {
	err := os.Remove(filepath.Join(s.root, filepath.Base(key)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
