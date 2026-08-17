package todo

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, params ListParams) (TodoPage, error) {
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return TodoPage{}, fmt.Errorf("%w: limit must be between 1 and 100", ErrInvalidInput)
	}
	skip := params.Skip
	if skip < 0 {
		return TodoPage{}, fmt.Errorf("%w: skip must be greater than or equal to zero", ErrInvalidInput)
	}

	page, err := s.repo.List(ctx, ListParams{Limit: limit, Skip: skip})
	if err != nil {
		return TodoPage{}, fmt.Errorf("list todos: %w", err)
	}
	return page, nil
}

func (s *Service) Get(ctx context.Context, id int) (Todo, error) {
	if id <= 0 {
		return Todo{}, fmt.Errorf("%w: todo id must be greater than zero", ErrInvalidInput)
	}

	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return Todo{}, fmt.Errorf("get todo %d: %w", id, err)
	}
	return t, nil
}

func (s *Service) GetByUser(ctx context.Context, userID int) (TodoPage, error) {
	if userID <= 0 {
		return TodoPage{}, fmt.Errorf("%w: user id must be greater than zero", ErrInvalidInput)
	}

	page, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return TodoPage{}, fmt.Errorf("get todos for user %d: %w", userID, err)
	}
	return page, nil
}
