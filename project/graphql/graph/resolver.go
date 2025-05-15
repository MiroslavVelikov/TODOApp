package graph

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"project/graphql/graph/model"
	"project/graphql/graph/utils"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type ServiceListInterface interface {
	CreateList(ctx context.Context, list model.List, requestCreator string) (*model.ListOutput, error)
	AddUserToList(ctx context.Context, listId, requestCreator string, newUser model.User) (string, error)
	UpdateListName(ctx context.Context, listId, requestCreator string, listUpdate model.List) (*model.ListOutput, error)
	DeleteList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error)
	RemoveUserFromList(ctx context.Context, listId, user, requestCreator string) (*model.UserOutput, error)
	GetList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error)
	GetLists(ctx context.Context, first *int32, after *string, requestCreator string) (*model.ListConnection, error)
	GetUserFromList(ctx context.Context, listId, user, requestCreator string) (*model.UserOutput, error)
	GetUsersFromList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error)
}

type ServiceTodoInterface interface {
	CreateTodo(ctx context.Context, listId, requestCreator string, todo *model.Todo) (*model.TodoOutput, error)
	UpdateTodo(ctx context.Context, listId, todoId, requestCreator string, todoUpdate *model.UpdateTodoInput) (*model.TodoOutput, error)
	DeleteTodo(ctx context.Context, listId, todoId, requestCreator string) (*model.TodoOutput, error)
	AssignUserToTodo(ctx context.Context, listId, todoId, requestCreator string) (string, error)
	ChangeTodoStatus(ctx context.Context, listId, todoId, requestCreator string) (string, error)
	GetTodoFromList(ctx context.Context, listId, todoId, requestCreator string) (*model.TodoOutput, error)
	GetTodosFromList(ctx context.Context, first *int32, after *string, listId, requestCreator string) (*model.TodoConnection, error)
}

type Resolver struct {
	listService ServiceListInterface
	todoService ServiceTodoInterface
}

func NewResolver(listService ServiceListInterface, todoService ServiceTodoInterface) *Resolver {
	return &Resolver{
		listService: listService,
		todoService: todoService,
	}
}

func HasAdminPermissionDirective(ctx context.Context, _ any, next graphql.Resolver) (res any, err error) {
	err = utils.ValidatePermission(ctx, utils.Admin)
	if err != nil {
		return nil, err
	}

	return next(ctx)
}

func HasWriterPermissionDirective(ctx context.Context, _ any, next graphql.Resolver) (res any, err error) {
	err = utils.ValidatePermission(ctx, utils.Writer)
	if err != nil {
		return nil, err
	}

	return next(ctx)
}

func HasReaderPermissionDirective(ctx context.Context, _ any, next graphql.Resolver) (res any, err error) {
	err = utils.ValidatePermission(ctx, utils.Reader)
	if err != nil {
		return nil, err
	}

	return next(ctx)
}
