package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"
	"project/graphql/graph/model"
	"project/graphql/graph/utils"
)

// CreateList is the resolver for the createList field.
func (r *mutationResolver) CreateList(ctx context.Context, list model.List) (*model.ListOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.CreateList(ctx, list, requestCreator)
}

// AddUserToList is the resolver for the addUser field.
func (r *mutationResolver) AddUserToList(ctx context.Context, listID string, user model.User) (string, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.AddUserToList(ctx, listID, requestCreator, user)
}

// CreateTodo is the resolver for the createTodo field.
func (r *mutationResolver) CreateTodo(ctx context.Context, listID string, todo *model.Todo) (*model.TodoOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.CreateTodo(ctx, listID, requestCreator, todo)
}

// UpdateListName is the resolver for the updateListName field.
func (r *mutationResolver) UpdateListName(ctx context.Context, listID string, input *model.List) (*model.ListOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.UpdateListName(ctx, listID, requestCreator, *input)
}

// UpdateTodo is the resolver for the updateTodo field.
func (r *mutationResolver) UpdateTodo(ctx context.Context, listID string, todoID string, todo *model.UpdateTodoInput) (*model.TodoOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.UpdateTodo(ctx, listID, todoID, requestCreator, todo)
}

// DeleteList is the resolver for the deleteList field.
func (r *mutationResolver) DeleteList(ctx context.Context, listID string) (*model.ListOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.DeleteList(ctx, listID, requestCreator)
}

// RemoveUserFromList is the resolver for the removeUser field.
func (r *mutationResolver) RemoveUserFromList(ctx context.Context, listID string, userID string) (*model.UserOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.RemoveUserFromList(ctx, listID, userID, requestCreator)
}

// DeleteTodo is the resolver for the deleteTodo field.
func (r *mutationResolver) DeleteTodo(ctx context.Context, listID string, todoID string) (*model.TodoOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.DeleteTodo(ctx, listID, todoID, requestCreator)
}

// AssignUserToTodo is the resolver for the assignUserToTodo field.
func (r *mutationResolver) AssignUserToTodo(ctx context.Context, listID string, todoID string) (string, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.AssignUserToTodo(ctx, listID, todoID, requestCreator)
}

// ChangeTodoStatus is the resolver for the changeTodoStatus field.
func (r *mutationResolver) ChangeTodoStatus(ctx context.Context, listID string, todoID string) (string, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.ChangeTodoStatus(ctx, listID, todoID, requestCreator)
}

// List is the resolver for the list field.
func (r *queryResolver) List(ctx context.Context, listID string) (*model.ListOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.GetList(ctx, listID, requestCreator)
}

// Lists is the resolver for the lists field.
func (r *queryResolver) Lists(ctx context.Context, first *int32, after *string) (*model.ListConnection, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.GetLists(ctx, first, after, requestCreator)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, listID string, userID string) (*model.UserOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.GetUserFromList(ctx, listID, userID, requestCreator)
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, listID string) (*model.ListOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.listService.GetUsersFromList(ctx, listID, requestCreator)
}

// Todo is the resolver for the todo field.
func (r *queryResolver) Todo(ctx context.Context, listID string, todoID string) (*model.TodoOutput, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.GetTodoFromList(ctx, listID, todoID, requestCreator)
}

// Todos is the resolver for the todos field.
func (r *queryResolver) Todos(ctx context.Context, listID string, first *int32, after *string) (*model.TodoConnection, error) {
	requestCreator := ctx.Value(utils.Username).(string)
	return r.todoService.GetTodosFromList(ctx, first, after, listID, requestCreator)
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
