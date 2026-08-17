package todo

import "context"

type Repository interface {
	List(ctx context.Context, params ListParams) (TodoPage, error)
	Get(ctx context.Context, id int) (Todo, error)
	GetByUser(ctx context.Context, userID int) (TodoPage, error)
}
