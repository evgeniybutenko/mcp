package todos

import (
	"context"
	"errors"
	"fmt"

	"mcp-example/internal/todo"
	"mcp-example/pkg/http"
)

type Client struct {
	client  *http.Client
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{client: http.New(), baseURL: baseURL}
}

func (c *Client) List(ctx context.Context, params todo.ListParams) (todo.TodoPage, error) {
	u := fmt.Sprintf("%s/todos?limit=%d&skip=%d", c.baseURL, params.Limit, params.Skip)

	var dto todoPageDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dto); err != nil {
		return todo.TodoPage{}, mapError(err)
	}

	return toTodoPage(dto), nil
}

func (c *Client) Get(ctx context.Context, id int) (todo.Todo, error) {
	u := fmt.Sprintf("%s/todos/%d", c.baseURL, id)

	var dto todoDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dto); err != nil {
		return todo.Todo{}, mapError(err)
	}

	return toTodo(dto), nil
}

func (c *Client) GetByUser(ctx context.Context, userID int) (todo.TodoPage, error) {
	u := fmt.Sprintf("%s/todos/user/%d", c.baseURL, userID)

	var dto todoPageDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dto); err != nil {
		return todo.TodoPage{}, mapError(err)
	}

	return toTodoPage(dto), nil
}

func toTodo(dto todoDTO) todo.Todo {
	return todo.Todo{
		ID:        dto.ID,
		Text:      dto.Todo,
		Completed: dto.Completed,
		UserID:    dto.UserID,
	}
}

func toTodoPage(dto todoPageDTO) todo.TodoPage {
	items := make([]todo.Todo, 0, len(dto.Todos))
	for _, t := range dto.Todos {
		items = append(items, toTodo(t))
	}
	return todo.TodoPage{
		Items: items,
		Total: dto.Total,
		Skip:  dto.Skip,
		Limit: dto.Limit,
	}
}

func mapError(err error) error {
	var httpErr *http.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.IsNotFound() {
			return fmt.Errorf("%w: %s", todo.ErrNotFound, httpErr.Status)
		}
		return fmt.Errorf("%w: %s", todo.ErrUpstream, httpErr.Status)
	}
	return err
}
