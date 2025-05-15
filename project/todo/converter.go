package todo

import (
	"project/structures"
)

type ServiceTodoConvertor struct{}

func NewServiceTodoConvertor() *ServiceTodoConvertor {
	return &ServiceTodoConvertor{}
}

func (s *ServiceTodoConvertor) ConvertTodoModelToOutput(todoModel *structures.TodoModel) *structures.TodoOutput {
	todoOutput := structures.TodoOutput{
		Id:          todoModel.Id,
		ListId:      todoModel.ListId,
		Name:        todoModel.Name,
		Description: todoModel.Description,
		Deadline:    todoModel.Deadline,
		Assignee:    todoModel.Assignee,
		Status:      todoModel.Status,
		Priority:    todoModel.Priority,
	}

	return &todoOutput
}

func (s *ServiceTodoConvertor) ConvertTodoModelToEntity(todoModel *structures.TodoModel) *structures.TodoEntity {
	todoEntity := structures.TodoEntity{
		Id:          todoModel.Id,
		ListId:      todoModel.ListId,
		Name:        todoModel.Name,
		Description: todoModel.Description,
		Deadline:    todoModel.Deadline,
		Assignee:    todoModel.Assignee,
		Priority:    todoModel.Priority,
		Status:      todoModel.Status,
	}

	return &todoEntity
}

type RepositoryTodoConvertor struct{}

func NewRepositoryTodoConvertor() *RepositoryTodoConvertor {
	return &RepositoryTodoConvertor{}
}

func (r *RepositoryTodoConvertor) ConvertEntityToModel(entity structures.TodoEntity) structures.TodoModel {
	return structures.TodoModel{
		Id:           entity.Id,
		Name:         entity.Name,
		ListId:       entity.ListId,
		Description:  entity.Description,
		Deadline:     entity.Deadline,
		CreationDate: entity.CreationDate,
		Assignee:     entity.Assignee,
		Priority:     entity.Priority,
		Status:       entity.Status,
	}
}

func (r *RepositoryTodoConvertor) ConvertEntitiesToModels(entities []structures.TodoEntity) []structures.TodoModel {
	models := make([]structures.TodoModel, len(entities))
	for i, e := range entities {
		model := r.ConvertEntityToModel(e)
		models[i] = model
	}

	return models
}
