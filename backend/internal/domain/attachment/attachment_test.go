package attachment

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"
)

type storageStub struct{ saved []byte }

func (s *storageStub) Save(_ context.Context, r io.Reader) (string, error) {
	s.saved, _ = io.ReadAll(r)
	return "key", nil
}
func (s *storageStub) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.saved)), nil
}
func (s *storageStub) Delete(context.Context, string) error { return nil }

type repositoryStub struct{ mime string }

func (r *repositoryStub) Upsert(_ context.Context, _, _, _, _, _, _, mime string, size int64) (Attachment, string, error) {
	r.mime = mime
	return Attachment{ID: "id", MimeType: mime, Size: size}, "", nil
}
func (r *repositoryStub) Get(context.Context, string, string) (Stored, error) { return Stored{}, nil }
func (r *repositoryStub) Delete(context.Context, string, string, string, string) (string, error) {
	return "", nil
}
func TestUploadPNG(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	repo := &repositoryStub{}
	store := &storageStub{}
	item, err := NewService(repo, store).Upload(context.Background(), "u", "j", "r", "c", "test.png", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if item.MimeType != "image/png" || repo.mime != "image/png" || !bytes.Equal(store.saved, data) {
		t.Fatalf("unexpected upload result")
	}
}
func TestUploadRejectsText(t *testing.T) {
	_, err := NewService(&repositoryStub{}, &storageStub{}).Upload(context.Background(), "u", "j", "r", "c", "file.txt", 4, strings.NewReader("text"))
	if err != ErrInvalidFile {
		t.Fatalf("expected ErrInvalidFile, got %v", err)
	}
}
