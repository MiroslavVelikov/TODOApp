package todo

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
	firstTodo                = 0
	failedToDeleteTodoErrMsg = "failed to delete todo"
)

//go:generate mockery --name ServiceConverterTodo --output=automock --with-expecter=true
type ServiceConverterTodo interface {
	ConvertResponseToTodoOutput(response []byte) (*model.TodoOutput, error)
	ConvertResponseToTodosOutputs(response []byte) ([]*model.TodoOutput, error)
}

//go:generate mockery --name RequestSenderInterface --output=automock --with-expecter=true
type RequestSenderInterface interface {
	SendRequest(requestType, route string, body any, headerData map[string]string, expectedStatus int) ([]byte, error, int)
}

type ServiceTodo struct {
	requestSender RequestSenderInterface
	converter     ServiceConverterTodo
	pageInfo      model.PageInfo
}

func NewServiceTodo(converter ServiceConverterTodo, requestSender *RequestSenderInterface) *ServiceTodo {
	if requestSender == nil {
		var reqSenderInterface RequestSenderInterface = utils.NewRequestSender()
		requestSender = &reqSenderInterface
	}

	return &ServiceTodo{
		requestSender: *requestSender,
		converter:     converter,
	}
}

func (st *ServiceTodo) CreateTodo(ctx context.Context, listId, requestCreator string, todo *model.Todo) (*model.TodoOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodPost, url, todo, headers, http.StatusCreated)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	todoOutput, err := st.converter.ConvertResponseToTodoOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*todoOutput)
	return todoOutput, nil
}

func (st *ServiceTodo) UpdateTodo(ctx context.Context, listId, todoId, requestCreator string, todoUpdate *model.UpdateTodoInput) (*model.TodoOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", listId, todoId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodPut, url, todoUpdate, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	todoOutput, err := st.converter.ConvertResponseToTodoOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*todoOutput)
	return todoOutput, nil
}

func (st *ServiceTodo) DeleteTodo(ctx context.Context, listId, todoId, requestCreator string) (*model.TodoOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", listId, todoId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodDelete, url, nil, headers, http.StatusOK)
	if err != nil {
		ok := strings.Contains(err.Error(), failedToDeleteTodoErrMsg)
		if ok {
			err = errors.New(fmt.Sprintf("not found todo (id: %s) in list with id: %s", todoId, listId))
			log.WithField(utils.Status, http.StatusNoContent).Error(err)
			return nil, errors.New("")
		}
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	todoOutput, err := st.converter.ConvertResponseToTodoOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*todoOutput)
	return todoOutput, nil
}

func (st *ServiceTodo) AssignUserToTodo(ctx context.Context, listId, todoId, requestCreator string) (string, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", listId, todoId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodPatch, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return "", err
	}

	strResult := string(result)
	log.WithField(utils.Status, status).Info(strResult)
	return strResult, nil
}

func (st *ServiceTodo) ChangeTodoStatus(ctx context.Context, listId, todoId, requestCreator string) (string, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s/status", listId, todoId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodPatch, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return "", err
	}

	strResult := string(result)
	log.WithField(utils.Status, status).Info(strResult)
	return strResult, nil
}

func (st *ServiceTodo) GetTodoFromList(ctx context.Context, listId, todoId, requestCreator string) (*model.TodoOutput, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", listId, todoId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	todoOutput, err := st.converter.ConvertResponseToTodoOutput(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	log.WithField(utils.Status, status).Info(*todoOutput)
	return todoOutput, nil
}

func (st *ServiceTodo) GetTodosFromList(ctx context.Context, first *int32, after *string, listId, requestCreator string) (*model.TodoConnection, error) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todos", listId)
	headers := map[string]string{
		utils.Username: requestCreator,
	}

	log := ctx.Value(utils.Logger).(*logrus.Entry)
	result, err, status := st.requestSender.SendRequest(http.MethodGet, url, nil, headers, http.StatusOK)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}

	todosOutputs, err := st.converter.ConvertResponseToTodosOutputs(result)
	if err != nil {
		log.WithField(utils.Status, http.StatusInternalServerError).Error(err)
		return nil, err
	}
	totalCount := int32(len(todosOutputs))

	if first == nil && after == nil {
		st.pageInfo = *new(model.PageInfo)
		return &model.TodoConnection{
			TotalCount: &totalCount,
			Todos:      todosOutputs,
			PageInfo:   &st.pageInfo,
		}, nil
	}

	if first == nil {
		first = &totalCount
	}

	var startPos int
	if after != nil || st.pageInfo.EndCursor != nil {
		var afterList *string
		if after != nil {
			afterList = after
		} else {
			afterList = st.pageInfo.EndCursor
		}

		startPos, err = utils.GetTodoPosition(*afterList, todosOutputs)
		if err != nil {
			log.WithField(utils.Status, http.StatusInternalServerError).Error(err.Error())
			st.pageInfo = *new(model.PageInfo)
			return nil, err
		} else if startPos == int(totalCount)-1 {
			err = errors.New("todo is out of range")
			log.WithField(utils.Status, http.StatusResetContent).Error(err)
			st.pageInfo = *new(model.PageInfo)
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
	pagedTodos := todosOutputs[startPos:pageScope]
	st.pageInfo.StartCursor = &pagedTodos[firstTodo].ID
	st.pageInfo.EndCursor = &pagedTodos[len(pagedTodos)-1].ID
	if pageScope < int(totalCount) {
		st.pageInfo.HasNextPage = true
	} else {
		st.pageInfo.HasNextPage = false
	}

	todoConnection := &model.TodoConnection{
		TotalCount: &totalCount,
		Todos:      pagedTodos,
		PageInfo:   &st.pageInfo,
	}

	log.WithField(utils.Status, status).Info("todos are successfully registered")
	return todoConnection, nil
}
