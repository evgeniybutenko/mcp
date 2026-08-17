package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/todo"
)

type todosListInput struct {
	Limit int `json:"limit" jsonschema:"optional maximum number of todos to return (1-100, default 20)"`
	Skip  int `json:"skip" jsonschema:"optional number of todos to skip (default 0)"`
}

type todosListOutput struct {
	Items []todoOutputItem `json:"items"`
	Total int              `json:"total"`
	Skip  int              `json:"skip"`
	Limit int              `json:"limit"`
}

type todoOutputItem struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	UserID    int    `json:"user_id"`
}

type todosGetInput struct {
	ID int `json:"id" jsonschema:"the todo ID to retrieve"`
}

type todosByUserInput struct {
	UserID int `json:"user_id" jsonschema:"the user ID to list todos for"`
}

func registerTodosTools(server *mcp.Server, svc *todo.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "todos_list",
		Description: "List todos with pagination.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in todosListInput) (*mcp.CallToolResult, todosListOutput, error) {
		page, err := svc.List(ctx, todo.ListParams{Limit: in.Limit, Skip: in.Skip})
		if err != nil {
			return nil, todosListOutput{}, err
		}
		return nil, toTodosListOutput(page), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todos_get",
		Description: "Get a single todo by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in todosGetInput) (*mcp.CallToolResult, todoOutputItem, error) {
		t, err := svc.Get(ctx, in.ID)
		if err != nil {
			return nil, todoOutputItem{}, err
		}
		return nil, toTodoOutput(t), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todos_by_user",
		Description: "Get todos belonging to a specific user.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in todosByUserInput) (*mcp.CallToolResult, todosListOutput, error) {
		page, err := svc.GetByUser(ctx, in.UserID)
		if err != nil {
			return nil, todosListOutput{}, err
		}
		return nil, toTodosListOutput(page), nil
	})
}

func toTodoOutput(t todo.Todo) todoOutputItem {
	return todoOutputItem{
		ID:        t.ID,
		Text:      t.Text,
		Completed: t.Completed,
		UserID:    t.UserID,
	}
}

func toTodosListOutput(page todo.TodoPage) todosListOutput {
	items := make([]todoOutputItem, 0, len(page.Items))
	for _, t := range page.Items {
		items = append(items, toTodoOutput(t))
	}
	return todosListOutput{
		Items: items,
		Total: page.Total,
		Skip:  page.Skip,
		Limit: page.Limit,
	}
}
