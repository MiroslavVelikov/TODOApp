package list

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"project/graphql/graph/model"
	"project/graphql/graph/utils"
	"strings"
)

const (
	firstList          = 0
	notFoundListErrMsg = "error not found list with id:"
)

//go:generate mockery --name ServiceConverterList --output=automock --with-expecter=true
type ServiceConverterList interface {
	ConvertResponseToListOutput(response []byte) (*model.ListOutput, error)
	ConvertListOutputToGQLListOutput(response *model.ListOutput) (*model.ListOutput, error)
	ConvertResponseToUserOutput(response []byte) (*model.UserOutput, error)
	ConvertResponseToListsOutputs(response []byte) ([]*model.ListOutput, error)
	ConvertResponseToTodosOutputs(response []byte) ([]*model.TodoOutput, error)
}

//go:generate mockery --name RequestSenderInterface --output=automock --with-expecter=true
type RequestSenderInterface interface {
	SendRequest(requestType, route string, body any, headerData map[string]string, expectedStatus int) ([]byte, error, int)
}

type ServiceList struct {
	requestSender RequestSenderInterface
	converter     ServiceConverterList
	pageInfo      model.PageInfo
}

func NewServiceList(converter ServiceConverterList, requestSender *RequestSenderInterface) *ServiceList {
	if requestSender == nil {
		var reqSenderInterface RequestSenderInterface = utils.NewRequestSender()
		requestSender = &reqSenderInterface
	}

	return &ServiceList{
		requestSender: *requestSender,
		converter:     converter,
	}
}

func (sl *ServiceList) CreateList(ctx context.Context, list model.List, requestCreator string) (*model.ListOutput, error) {
	url := utils.BaseUrl + utils.BasePath + "/list"
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodPost, url, list, headers, http.StatusCreated)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listOutput, err := sl.converter.ConvertResponseToListOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*listOutput)
	return listOutput, nil
}

func (sl *ServiceList) AddUserToList(ctx context.Context, listId, requestCreator string, newUser model.User) (string, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodPost, url, newUser, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return "", err
	}

	strResult := string(result)
	log.WithField(utils.Status, status).Info(strResult)
	return strResult, nil
}

func (sl *ServiceList) UpdateListName(ctx context.Context, listId, requestCreator string, listUpdate model.List) (*model.ListOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodPut, url, listUpdate, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listOutput, err := sl.converter.ConvertResponseToListOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*listOutput)
	return listOutput, nil
}

func (sl *ServiceList) DeleteList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodDelete, url, nil, headers, http.StatusOK)
	if err != nil {
		ok := strings.Contains(err.Error(), notFoundListErrMsg)
		if ok {
			err = errors.New(fmt.Sprintf("not found list with id: %s", listId))
			log.WithField(utils.Status, http.StatusNoContent).Error(err)
			return nil, errors.New("")
		}

		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listOutput, err := sl.converter.ConvertResponseToListOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*listOutput)
	return listOutput, nil
}

func (sl *ServiceList) RemoveUserFromList(ctx context.Context, listId, user, requestCreator string) (*model.UserOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users/%s", listId, user)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodDelete, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	userOutput, err := sl.converter.ConvertResponseToUserOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*userOutput)
	return userOutput, nil
}

func (sl *ServiceList) GetList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listOutput, err := sl.converter.ConvertResponseToListOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	url = fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todos", listId)
	result, err, status = sl.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}
	todosOutputs, err := sl.converter.ConvertResponseToTodosOutputs(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}
	listOutput.Todos = todosOutputs

	log.WithField(utils.Status, status).Info(*listOutput)
	return listOutput, nil
}

func (sl *ServiceList) GetLists(ctx context.Context, first *int32, after *string, requestCreator string) (*model.ListConnection, error) {
	url := utils.BaseUrl + utils.BasePath + "/list"
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listsOutputs, err := sl.converter.ConvertResponseToListsOutputs(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}
	totalCount := int32(len(listsOutputs))

	if first == nil && after == nil {
		sl.pageInfo = *new(model.PageInfo)
		return &model.ListConnection{
			TotalCount: &totalCount,
			Lists:      listsOutputs,
			PageInfo:   &sl.pageInfo,
		}, nil
	}

	if first == nil {
		first = &totalCount
	}

	var startPos int
	if after != nil || sl.pageInfo.EndCursor != nil {
		var afterList *string
		if after != nil {
			afterList = after
		} else {
			afterList = sl.pageInfo.EndCursor
		}

		startPos, err = utils.GetListPosition(*afterList, listsOutputs)
		if err != nil {
			log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
			sl.pageInfo = *new(model.PageInfo)
			return nil, err
		} else if startPos == int(totalCount)-1 {
			err = errors.New("lists is out of range")
			log.WithField(utils.Status, http.StatusResetContent).Error(err)
			sl.pageInfo = *new(model.PageInfo)
			return nil, err
		}

		startPos++
	} else {
		startPos = 0
	}

	pageScope := startPos + int(*first)
	if pageScope > int(totalCount) {
		pageScope = int(totalCount)
	}
	pagedLists := listsOutputs[startPos:pageScope]
	sl.pageInfo.StartCursor = &pagedLists[firstList].ID
	sl.pageInfo.EndCursor = &pagedLists[len(pagedLists)-1].ID
	if pageScope < int(totalCount) {
		sl.pageInfo.HasNextPage = true
	} else {
		sl.pageInfo.HasNextPage = false
	}

	listConnection := &model.ListConnection{
		TotalCount: &totalCount,
		Lists:      pagedLists,
		PageInfo:   &sl.pageInfo,
	}

	log.WithField(utils.Status, status).Info("lists are successfully retrieved")
	return listConnection, nil
}

func (sl *ServiceList) GetUserFromList(ctx context.Context, listId, user, requestCreator string) (*model.UserOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users/%s", listId, user)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	userOutput, err := sl.converter.ConvertResponseToUserOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*userOutput)
	return userOutput, nil
}

func (sl *ServiceList) GetUsersFromList(ctx context.Context, listId, requestCreator string) (*model.ListOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := sl.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	listOutput, err := sl.converter.ConvertResponseToListOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*listOutput)
	return listOutput, nil
}
