package list

import (
	"encoding/json"
	"project/graphql/graph/model"
	restStructures "project/structures"
)

type ConverterList struct{}

func NewListConverter() *ConverterList {
	return &ConverterList{}
}

func (cl *ConverterList) ConvertResponseToListOutput(response []byte) (*model.ListOutput, error) {
	var listOutput model.ListOutput
	err := json.Unmarshal(response, &listOutput)
	if err != nil {
		return nil, err
	}

	return &listOutput, nil
}

func (cl *ConverterList) ConvertResponseToUserOutput(response []byte) (*model.UserOutput, error) {
	var userOutputResponse restStructures.UserOutput
	err := json.Unmarshal(response, &userOutputResponse)
	if err != nil {
		return nil, err
	}

	userOutput := &model.UserOutput{
		ListID:   userOutputResponse.ListId.String(),
		ListName: userOutputResponse.ListName,
		Username: userOutputResponse.Username,
		IsOwner:  userOutputResponse.IsOwner,
	}
	return userOutput, nil
}

func (cl *ConverterList) ConvertResponseToListsOutputs(response []byte) ([]*model.ListOutput, error) {
	var listsOutputs []*model.ListOutput
	err := json.Unmarshal(response, &listsOutputs)
	if err != nil {
		return nil, err
	}

	return listsOutputs, nil
}

func (cl *ConverterList) ConvertResponseToTodosOutputs(response []byte) ([]*model.TodoOutput, error) {
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
