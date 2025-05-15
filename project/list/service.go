package list

import (
	"context"
	"github.com/google/uuid"
	"project/structures"
)

//go:generate mockery --name RepositoryList --output=automock --with-expecter=true
type RepositoryList interface {
	GetListById(ctx context.Context, listId uuid.UUID) (*structures.ListModel, error)
	GetAllLists(ctx context.Context) []*structures.ListModel
	GetListOwner(ctx context.Context, listId uuid.UUID) (*structures.UserModel, error)
	GetUserFromListById(ctx context.Context, listId uuid.UUID, username string) (*structures.UserModel, error)
	CreateList(ctx context.Context, entityList structures.ListEntity, entityUser structures.ListUserEntity) error
	AddUserToList(ctx context.Context, entityUser structures.ListUserEntity) error
	DeleteList(ctx context.Context, listId uuid.UUID) (*structures.ListModel, error)
	RemoveUserUserFromList(ctx context.Context, entityUser structures.ListUserEntity) (*structures.UserModel, error)
	UpdateList(ctx context.Context, listId uuid.UUID, newListName string) (*structures.ListModel, error)
	CheckIfListExists(ctx context.Context, listId uuid.UUID) bool
	ContainsUserInList(ctx context.Context, listId uuid.UUID, username string) bool
}

type ServiceListImpl struct {
	repo      RepositoryList
	converter ServiceConvertorList
}

func NewServiceList(repo RepositoryList, converter ServiceConvertorList) *ServiceListImpl {
	return &ServiceListImpl{repo: repo, converter: converter}
}

func (s *ServiceListImpl) GetListById(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error) {
	listModel, err := s.repo.GetListById(ctx, listId)
	if err != nil {
		return nil, err
	}

	return s.converter.ConvertListModelToListUserOutput(listModel), nil
}

func (s *ServiceListImpl) GetAllLists(ctx context.Context) []*structures.ListOutput {
	result := s.repo.GetAllLists(ctx)
	return s.converter.ConvertListModelsToOutputs(result)
}

func (s *ServiceListImpl) GetUserFromListById(ctx context.Context, listId uuid.UUID, username string) (*structures.UserOutput, error) {
	userModel, err := s.repo.GetUserFromListById(ctx, listId, username)
	if err != nil {
		return nil, err
	}
	return s.converter.ConvertUserModelToUserOutput(userModel), err
}

func (s *ServiceListImpl) GetUsersFromListById(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error) {
	listModel, err := s.repo.GetListById(ctx, listId)
	if err != nil {
		return nil, err
	}

	return s.converter.ConvertListModelToListUserOutput(listModel), nil
}

func (s *ServiceListImpl) CreateList(ctx context.Context, listName, username string) (*structures.ListOutput, error) {
	listModel := structures.ListModel{
		Id:    uuid.New(),
		Name:  listName,
		Owner: username,
		Users: []string{username},
	}

	listEntity, listUserEntity := s.converter.ConvertListModelToEntities(&listModel)
	err := s.repo.CreateList(ctx, *listEntity, *listUserEntity)
	if err != nil {
		return nil, err
	}

	return s.converter.ConvertListModelToOutput(&listModel), nil
}

func (s *ServiceListImpl) AddUserToList(ctx context.Context, listId uuid.UUID, username string) error {
	entityUser := structures.ListUserEntity{
		Username: username,
		ListId:   listId,
		IsOwner:  false,
	}

	return s.repo.AddUserToList(ctx, entityUser)
}

func (s *ServiceListImpl) DeleteList(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error) {
	deletedList, err := s.repo.DeleteList(ctx, listId)
	if err != nil {
		return nil, err
	}

	return s.converter.ConvertListModelToUserOutput(deletedList), nil
}

func (s *ServiceListImpl) RemoveUserFromList(ctx context.Context, listId uuid.UUID, username string) (*structures.UserOutput, error) {
	owner, err := s.repo.GetListOwner(ctx, listId)
	if err != nil {
		return nil, err
	}

	isOwner := owner.Username == username
	entityUser := structures.ListUserEntity{
		Username: username,
		ListId:   listId,
		IsOwner:  isOwner,
	}

	removedUser, err := s.repo.RemoveUserUserFromList(ctx, entityUser)
	if err != nil {
		return nil, err
	}
	return s.converter.ConvertUserModelToUserOutput(removedUser), nil
}

func (s *ServiceListImpl) UpdateList(ctx context.Context, listId uuid.UUID, newListName string) (*structures.ListOutput, error) {
	updatedList, err := s.repo.UpdateList(ctx, listId, newListName)
	if err != nil {
		return nil, err
	}

	return s.converter.ConvertListModelToOutput(updatedList), nil
}

func (s *ServiceListImpl) CheckIfListExistsInList(ctx context.Context, listId uuid.UUID) bool {
	return s.repo.CheckIfListExists(ctx, listId)
}

func (s *ServiceListImpl) ContainUserInList(ctx context.Context, listId uuid.UUID, username string) bool {
	return s.repo.ContainsUserInList(ctx, listId, username)
}
