package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/message/domain"

	"github.com/google/uuid"
)

type searchRepo struct {
	domain.MessageRepository
	search func(ctx context.Context) ([]*domain.Message, error)
}

func (r *searchRepo) Search(ctx context.Context, _ int, _ string, _ *uuid.UUID, _ int) ([]*domain.Message, error) {
	return r.search(ctx)
}

func TestSearchTimesOutAsTooBroad(t *testing.T) {
	repo := &searchRepo{search: func(ctx context.Context) ([]*domain.Message, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	uc := NewSearchMessagesUseCase(repo)
	uc.timeout = 20 * time.Millisecond
	if _, err := uc.Execute(context.Background(), 1, "hola", "", 10); !errors.Is(err, domain.ErrSearchTooBroad) {
		t.Fatalf("got %v, want ErrSearchTooBroad", err)
	}
}

func TestSearchRejectsWhenSaturated(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, maxConcurrentSearches)
	repo := &searchRepo{search: func(ctx context.Context) ([]*domain.Message, error) {
		started <- struct{}{}
		<-release
		return nil, nil
	}}
	uc := NewSearchMessagesUseCase(repo)
	done := make(chan error, maxConcurrentSearches)
	for i := 0; i < maxConcurrentSearches; i++ {
		go func() { _, err := uc.Execute(context.Background(), 1, "hola", "", 10); done <- err }()
	}
	for i := 0; i < maxConcurrentSearches; i++ {
		<-started
	}
	if _, err := uc.Execute(context.Background(), 1, "hola", "", 10); !errors.Is(err, domain.ErrSearchBusy) {
		t.Fatalf("got %v, want ErrSearchBusy", err)
	}
	close(release)
	for i := 0; i < maxConcurrentSearches; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if _, err := uc.Execute(context.Background(), 1, "hola", "", 10); err != nil {
		t.Fatalf("after release: %v", err)
	}
}

func TestSearchCallerCancelIsNotTooBroad(t *testing.T) {
	repo := &searchRepo{search: func(ctx context.Context) ([]*domain.Message, error) { return nil, ctx.Err() }}
	uc := NewSearchMessagesUseCase(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := uc.Execute(ctx, 1, "hola", "", 10); errors.Is(err, domain.ErrSearchTooBroad) {
		t.Fatal("caller cancellation must not be reported as too broad")
	}
}
