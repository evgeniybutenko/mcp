package todos

type todoDTO struct {
	ID        int    `json:"id"`
	Todo      string `json:"todo"`
	Completed bool   `json:"completed"`
	UserID    int    `json:"userId"`
}

type todoPageDTO struct {
	Todos []todoDTO `json:"todos"`
	Total int       `json:"total"`
	Skip  int       `json:"skip"`
	Limit int       `json:"limit"`
}
