package todo

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	listFunc    func(context.Context, ListParams) (TodoPage, error)
	getFunc     func(context.Context, int) (Todo, error)
	getByUserFn func(context.Context, int) (TodoPage, error)
}

func (f *fakeRepository) List(ctx context.Context, params ListParams) (TodoPage, error) {
	if f.listFunc == nil {
		return TodoPage{}, nil
	}
	return f.listFunc(ctx, params)
}

func (f *fakeRepository) Get(ctx context.Context, id int) (Todo, error) {
	if f.getFunc == nil {
		return Todo{}, nil
	}
	return f.getFunc(ctx, id)
}

func (f *fakeRepository) GetByUser(ctx context.Context, userID int) (TodoPage, error) {
	if f.getByUserFn == nil {
		return TodoPage{}, nil
	}
	return f.getByUserFn(ctx, userID)
}

func TestList_Valid(t *testing.T) {
	svc := NewService(&fakeRepository{
		listFunc: func(_ context.Context, params ListParams) (TodoPage, error) {
			if params.Limit != 20 {
				t.Errorf("expected limit 20, got %d", params.Limit)
			}
			if params.Skip != 0 {
				t.Errorf("expected skip 0, got %d", params.Skip)
			}
			return TodoPage{
				Items: []Todo{{ID: 1, Text: "test", Completed: false, UserID: 5}},
				Total: 1, Skip: 0, Limit: 20,
			}, nil
		},
	})

	page, err := svc.List(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
	if page.Items[0].ID != 1 {
		t.Errorf("expected item ID 1, got %d", page.Items[0].ID)
	}
}

func TestList_DefaultPagination(t *testing.T) {
	svc := NewService(&fakeRepository{
		listFunc: func(_ context.Context, params ListParams) (TodoPage, error) {
			if params.Limit != 20 {
				t.Errorf("default limit should be 20, got %d", params.Limit)
			}
			if params.Skip != 0 {
				t.Errorf("default skip should be 0, got %d", params.Skip)
			}
			return TodoPage{Limit: 20}, nil
		},
	})

	_, err := svc.List(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_LimitTooHigh(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.List(context.Background(), ListParams{Limit: 101})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestList_LimitTooLow(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.List(context.Background(), ListParams{Limit: 0, Skip: 0})
	if err != nil {
		t.Fatalf("limit=0 should default to 20, got error: %v", err)
	}

	_, err = svc.List(context.Background(), ListParams{Limit: -1})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for limit<1, got %v", err)
	}
}

func TestList_NegativeSkip(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.List(context.Background(), ListParams{Skip: -1})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGet_Valid(t *testing.T) {
	svc := NewService(&fakeRepository{
		getFunc: func(_ context.Context, id int) (Todo, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return Todo{ID: 1, Text: "test"}, nil
		},
	})

	td, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.ID != 1 {
		t.Errorf("expected ID 1, got %d", td.ID)
	}
}

func TestGet_InvalidID(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.Get(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	_, err = svc.Get(context.Background(), -1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetByUser_Valid(t *testing.T) {
	svc := NewService(&fakeRepository{
		getByUserFn: func(_ context.Context, userID int) (TodoPage, error) {
			if userID != 5 {
				t.Errorf("expected userID 5, got %d", userID)
			}
			return TodoPage{Items: []Todo{{ID: 1, UserID: 5}}, Total: 1}, nil
		},
	})

	page, err := svc.GetByUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
}

func TestGetByUser_InvalidUserID(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.GetByUser(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestList_RepoError(t *testing.T) {
	repoErr := errors.New("repo failure")
	svc := NewService(&fakeRepository{
		listFunc: func(context.Context, ListParams) (TodoPage, error) {
			return TodoPage{}, repoErr
		},
	})

	_, err := svc.List(context.Background(), ListParams{})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestGet_RepoError(t *testing.T) {
	repoErr := errors.New("repo failure")
	svc := NewService(&fakeRepository{
		getFunc: func(context.Context, int) (Todo, error) {
			return Todo{}, repoErr
		},
	})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
