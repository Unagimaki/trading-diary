package journal

import (
	"context"
	"testing"
)

type repositoryStub struct{ savedName string }

func (r *repositoryStub) List(context.Context, string) ([]Journal, error)      { return nil, nil }
func (r *repositoryStub) Get(context.Context, string, string) (Journal, error) { return Journal{}, nil }
func (r *repositoryStub) Delete(context.Context, string, string) error         { return nil }
func (r *repositoryStub) Create(_ context.Context, _ string, name string) (Journal, error) {
	r.savedName = name
	return Journal{Name: name}, nil
}
func (r *repositoryStub) Rename(context.Context, string, string, string) (Journal, error) {
	return Journal{}, nil
}

func TestCreateNormalizesName(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)
	_, err := service.Create(context.Background(), "user", "  Сделки  ")
	if err != nil {
		t.Fatal(err)
	}
	if repo.savedName != "Сделки" {
		t.Fatalf("unexpected name %q", repo.savedName)
	}
}
func TestCreateRejectsEmptyName(t *testing.T) {
	_, err := NewService(&repositoryStub{}).Create(context.Background(), "user", "   ")
	if err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}
