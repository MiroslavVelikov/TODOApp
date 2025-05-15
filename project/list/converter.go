package list

import (
	"project/structures"
)

type ServiceConvertorList struct{}

func NewServiceListConvertor() *ServiceConvertorList {
	return &ServiceConvertorList{}
}

func (s *ServiceConvertorList) ConvertListModelToOutput(listModel *structures.ListModel) *structures.ListOutput {
	return &structures.ListOutput{
		Id:    listModel.Id,
		Name:  listModel.Name,
		Owner: listModel.Owner,
	}
}

func (s *ServiceConvertorList) ConvertListModelToUserOutput(listModel *structures.ListModel) *structures.ListUserOutput {
	return &structures.ListUserOutput{
		Id:    listModel.Id,
		Name:  listModel.Name,
		Owner: listModel.Owner,
		Users: listModel.Users,
	}
}

func (s *ServiceConvertorList) ConvertListModelToEntities(listModel *structures.ListModel) (*structures.ListEntity, *structures.ListUserEntity) {
	listEntity := &structures.ListEntity{
		Id:   listModel.Id,
		Name: listModel.Name,
	}

	listUserEntity := &structures.ListUserEntity{
		ListId:   listModel.Id,
		Username: listModel.Owner,
		IsOwner:  true,
	}

	return listEntity, listUserEntity
}

func (s *ServiceConvertorList) ConvertUserModelToUserOutput(userModel *structures.UserModel) *structures.UserOutput {
	return &structures.UserOutput{
		ListId:   userModel.ListId,
		ListName: userModel.ListName,
		Username: userModel.Username,
		IsOwner:  userModel.IsOwner,
	}
}

func (s *ServiceConvertorList) ConvertListModelToUserOutputs(listModel *structures.ListModel) []*structures.UserOutput {
	outputs := make([]*structures.UserOutput, len(listModel.Users))
	for i, user := range listModel.Users {
		isOwner := user == listModel.Owner
		userOutput := structures.UserOutput{
			ListId:   listModel.Id,
			ListName: listModel.Name,
			Username: user,
			IsOwner:  isOwner,
		}
		outputs[i] = &userOutput
	}

	return outputs
}

func (s *ServiceConvertorList) ConvertListModelToListUserOutput(listModel *structures.ListModel) *structures.ListUserOutput {
	output := structures.ListUserOutput{
		Id:    listModel.Id,
		Name:  listModel.Name,
		Owner: listModel.Owner,
		Users: listModel.Users,
	}
	return &output
}

func (s *ServiceConvertorList) ConvertUserModelToOutput(userModel *structures.UserModel) *structures.UserOutput {
	return &structures.UserOutput{
		ListId:   userModel.ListId,
		ListName: userModel.ListName,
		Username: userModel.Username,
		IsOwner:  userModel.IsOwner,
	}
}

func (s *ServiceConvertorList) ConvertListModelsToOutputs(listModels []*structures.ListModel) []*structures.ListOutput {
	outputs := make([]*structures.ListOutput, len(listModels))
	for i, listModel := range listModels {
		outputs[i] = &structures.ListOutput{
			Id:    listModel.Id,
			Name:  listModel.Name,
			Owner: listModel.Owner,
		}
	}

	return outputs
}

type RepositoryConvertorList struct{}

func NewRepositoryListConvertor() *RepositoryConvertorList {
	return &RepositoryConvertorList{}
}

func (r *RepositoryConvertorList) ConvertEntitiesToModel(listEntity *structures.ListEntity, usernames []string, owner string) *structures.ListModel {
	return &structures.ListModel{
		Id:           listEntity.Id,
		Name:         listEntity.Name,
		CreationDate: listEntity.CreatedAt,
		Owner:        owner,
		Users:        usernames,
	}
}

func (r *RepositoryConvertorList) ConvertUserEntityToModel(userEntity structures.ListUserEntity, listName string) *structures.UserModel {
	return &structures.UserModel{
		ListId:   userEntity.ListId,
		ListName: listName,
		Username: userEntity.Username,
		IsOwner:  userEntity.IsOwner,
	}
}
