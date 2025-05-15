package todo

import (
	"encoding/json"
	"project/graphql/graph/model"
	restStructures "project/structures"
)

type ConverterTodo struct{}

func NewTodoConverter() *ConverterTodo {
	return &ConverterTodo{}
}

func (ct *ConverterTodo) ConvertResponseToTodoOutput(response []byte) (*model.TodoOutput, error) {
	var todoOutputResponse restStructures.TodoOutput
	err := json.Unmarshal(response, &todoOutputResponse)
	if err != nil {
		return nil, err
	}

	todoOutput := &model.TodoOutput{
		ID:          todoOutputResponse.Id.String(),
		ListID:      todoOutputResponse.ListId.String(),
		Name:        todoOutputResponse.Name,
		Description: todoOutputResponse.Description,
		Deadline:    todoOutputResponse.Deadline,
		Assignee:    todoOutputResponse.Assignee,
		Status:      todoOutputResponse.Status,
		Priority:    todoOutputResponse.Priority,
	}

	return todoOutput, nil
}

func (ct *ConverterTodo) ConvertResponseToTodosOutputs(response []byte) ([]*model.TodoOutput, error) {
	var todosOutputsResponse []*restStructures.TodoOutput
	err := json.Unmarshal(response, &todosOutputsResponse)
	if err != nil {
		return nil, err
	}

	todosOutputs := make([]*model.TodoOutput, len(todosOutputsResponse))
	for i, outputResponse := range todosOutputsResponse {
		todosOutputs[i] = &model.TodoOutput{
			ID:          outputResponse.Id.String(),
			ListID:      outputResponse.ListId.String(),
			Name:        outputResponse.Name,
			Description: outputResponse.Description,
			Deadline:    outputResponse.Deadline,
			Assignee:    outputResponse.Assignee,
			Status:      outputResponse.Status,
			Priority:    outputResponse.Priority,
		}
	}
	return todosOutputs, nil
}
