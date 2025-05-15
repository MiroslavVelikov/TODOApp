package todo

import (
	"context"
	"github.com/google/uuid"
	"project/structures"
	"project/utils"
)

//go:generate mockery --name RepositoryTodo --output=automock --with-expecter=true
type RepositoryTodo interface {
	GetTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoModel, error)
	GetAllTasks(ctx context.Context, listId uuid.UUID) []structures.TodoModel
	CreateTodo(ctx context.Context, newTask structures.TodoEntity) error
	DeleteTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoModel, error)
	UpdateTodo(ctx context.Context, updatedTask structures.TodoEntity, listId uuid.UUID) (*structures.TodoModel, error)
	AssignTodoToUser(ctx context.Context, todoId, listId uuid.UUID, username string) error
	ChangeTodoStatus(ctx context.Context, todoId, listId uuid.UUID) error
	CheckIfListContainsTodo(ctx context.Context, listId, todoId uuid.UUID) bool
	GetTodoAssignee(ctx context.Context, todoId uuid.UUID) string
}

type ServiceTodoImpl struct {
	repo      RepositoryTodo
	convertor ServiceTodoConvertor
}

func NewServiceTodo(repo RepositoryTodo, convertor ServiceTodoConvertor) *ServiceTodoImpl {
	return &ServiceTodoImpl{repo: repo, convertor: convertor}
}

func (s *ServiceTodoImpl) GetTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoOutput, error) {
	todoModel, err := s.repo.GetTodo(ctx, todoId, listId)
	if err != nil {
		return nil, err
	}

	return s.convertor.ConvertTodoModelToOutput(todoModel), nil
}

func (s *ServiceTodoImpl) GetAllTasks(ctx context.Context, listId uuid.UUID) []structures.TodoOutput {
	todoModels := s.repo.GetAllTasks(ctx, listId)
	result := make([]structures.TodoOutput, len(todoModels))
	for i, model := range todoModels {
		result[i] = *s.convertor.ConvertTodoModelToOutput(&model)
	}

	return result
}

func (s *ServiceTodoImpl) CreateTodo(ctx context.Context, input structures.TodoInput, listId uuid.UUID) (*structures.TodoOutput, error) {
	todoModel := structures.TodoModel{
		Id:          uuid.New(),
		ListId:      listId,
		Name:        input.Name,
		Description: input.Description,
		Deadline:    input.Deadline,
		Assignee:    "",
		Status:      utils.NotAssigned,
		Priority:    input.Priority,
	}

	err := s.repo.CreateTodo(ctx, *s.convertor.ConvertTodoModelToEntity(&todoModel))
	if err != nil {
		return nil, err
	}

	return s.convertor.ConvertTodoModelToOutput(&todoModel), nil
}

func (s *ServiceTodoImpl) DeleteTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoOutput, error) {
	deletedTodoModel, err := s.repo.DeleteTodo(ctx, todoId, listId)
	if err != nil {
		return nil, err
	}

	return s.convertor.ConvertTodoModelToOutput(deletedTodoModel), nil
}

func (s *ServiceTodoImpl) UpdateTodo(ctx context.Context, todoId, listId uuid.UUID, input structures.TodoInput) (*structures.TodoOutput, error) {
	todoModel := structures.TodoModel{
		Id:          todoId,
		Name:        input.Name,
		Description: input.Description,
		Deadline:    input.Deadline,
		Priority:    input.Priority,
	}

	todoUpdated, err := s.repo.UpdateTodo(ctx, *s.convertor.ConvertTodoModelToEntity(&todoModel), listId)
	if err != nil {
		return nil, err
	}
	todoUpdateOutput := s.convertor.ConvertTodoModelToOutput(todoUpdated)

	return todoUpdateOutput, nil
}

func (s *ServiceTodoImpl) AssignUserToTodo(ctx context.Context, todoId, listId uuid.UUID, username string) error {
	return s.repo.AssignTodoToUser(ctx, todoId, listId, username)
}

func (s *ServiceTodoImpl) ChangeTodoStatus(ctx context.Context, todoId, listId uuid.UUID) error {
	return s.repo.ChangeTodoStatus(ctx, todoId, listId)
}

func (s *ServiceTodoImpl) CheckIfListContainsTodo(ctx context.Context, todoId, listId uuid.UUID) bool {
	return s.repo.CheckIfListContainsTodo(ctx, todoId, listId)
}

func (s *ServiceTodoImpl) GetTodoAssignee(ctx context.Context, todoId uuid.UUID) string {
	return s.repo.GetTodoAssignee(ctx, todoId)
}
