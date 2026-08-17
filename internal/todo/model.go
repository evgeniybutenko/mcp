package todo

type Todo struct {
	ID        int
	Text      string
	Completed bool
	UserID    int
}

type TodoPage struct {
	Items []Todo
	Total int
	Skip  int
	Limit int
}

type ListParams struct {
	Limit int
	Skip  int
}
