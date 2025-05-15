package structures

import (
	"github.com/google/uuid"
	"time"
)

// For Resolver
type TodoInput struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Priority    string    `json:"priority"`
}

type TodoOutput struct {
	Id          uuid.UUID `json:"id"`
	ListId      uuid.UUID `json:"list_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Assignee    string    `json:"assignee"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
}

// For Service
type TodoModel struct {
	Id           uuid.UUID
	ListId       uuid.UUID
	Name         string
	Description  string
	Deadline     time.Time
	CreationDate time.Time
	Assignee     string
	Username     string
	Status       string
	Priority     string
}

// For Repository
type TodoEntity struct {
	Id           uuid.UUID `db:"id"`
	ListId       uuid.UUID `db:"list_id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	Deadline     time.Time `db:"deadline"`
	CreationDate time.Time `db:"created_at"`
	Assignee     string    `db:"assignee"`
	Status       string    `db:"status"`
	Priority     string    `db:"priority"`
}
