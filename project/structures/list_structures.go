package structures

import (
	"github.com/google/uuid"
	"time"
)

// For Resolver
type ListInput struct {
	Name string `json:"name"`
}

type ListUserInput struct {
	Username string `json:"username"`
}

// For Service
type ListModel struct {
	Id           uuid.UUID
	Name         string
	CreationDate time.Time
	Owner        string
	Users        []string
}

type UserModel struct {
	ListId   uuid.UUID
	ListName string
	Username string
	IsOwner  bool
}

// For Repository
type ListEntity struct {
	Id        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type ListUserEntity struct {
	ListId   uuid.UUID `db:"list_id"`
	Username string    `db:"username"`
	IsOwner  bool      `db:"is_owner"`
}

// For Resolver
type ListOutput struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Owner string    `json:"owner"`
}

type ListUserOutput struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Owner string    `json:"owner"`
	Users []string  `json:"users"`
}

type UserOutput struct {
	ListId   uuid.UUID `json:"list_id"`
	ListName string    `json:"list_name"`
	Username string    `json:"username"`
	IsOwner  bool      `json:"is_owner"`
}
