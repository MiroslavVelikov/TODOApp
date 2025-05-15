package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"project/structures"
	"project/utils"
	"regexp"
	"strings"
	"time"
)

const (
	listId                 = "listId"
	todoId                 = "todoId"
	username               = "userId"
	assigningErrorMsg      = "error assigning"
	changingStatusErrorMsg = "error changing status"
)

//go:generate mockery --name ServiceTodo --output=automock --with-expecter=true
type ServiceTodo interface {
	GetTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoOutput, error)
	GetAllTasks(ctx context.Context, listId uuid.UUID) []structures.TodoOutput
	CreateTodo(ctx context.Context, todoInput structures.TodoInput, listId uuid.UUID) (*structures.TodoOutput, error)
	DeleteTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoOutput, error)
	UpdateTodo(ctx context.Context, todoId, listId uuid.UUID, todoUpdate structures.TodoInput) (*structures.TodoOutput, error)
	AssignUserToTodo(ctx context.Context, todoId, listId uuid.UUID, username string) error
	ChangeTodoStatus(ctx context.Context, todoId, listId uuid.UUID) error
	CheckIfListContainsTodo(ctx context.Context, todoId, listId uuid.UUID) bool
	GetTodoAssignee(ctx context.Context, todoId uuid.UUID) string
}

type ResolverTodo struct {
	service ServiceTodo
}

func NewResolverTodo(service ServiceTodo) *ResolverTodo {
	return &ResolverTodo{
		service: service,
	}
}

func NewResolverWithService(service ServiceTodo) *ResolverTodo {
	return &ResolverTodo{
		service: service,
	}
}

func (r *ResolverTodo) GetTodo(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	vars := mux.Vars(req)
	todoId, err := utils.GetID(vars, todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	output, err := r.service.GetTodo(ctx, *todoId, *listId)
	if err != nil {
		isGetError, regErr := regexp.MatchString(utils.GetErrorMsg, err.Error())
		if isGetError && regErr == nil {
			w.WriteHeader(http.StatusNotFound)
			utils.ResponseHandling(req, w, err.Error())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("failed to get task with id: %s from list with id: %s", todoId, listId)
		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("Success getting todo with id: %s", todoId))
	utils.ResponseHandling(req, w, output)
}

func (r *ResolverTodo) GetAllTasks(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	vars := mux.Vars(req)
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	result := r.service.GetAllTasks(ctx, *listId)

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success getting all tasks form list with id: %s", *listId))
	utils.ResponseHandling(req, w, result)
}

func (r *ResolverTodo) validateTodo(input structures.TodoInput) bool {
	var defaultTime time.Time
	if input.Name == "" || input.Description == "" ||
		input.Deadline == defaultTime || input.Priority == utils.UnknownPriority {
		return false
	}

	return true
}

func (r *ResolverTodo) CreateTodo(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	var input structures.TodoInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := "error decoding body"
		utils.ResponseHandling(req, w, msg)
		return
	}
	if !r.validateTodo(input) {
		w.WriteHeader(http.StatusBadRequest)
		msg := "missing required field(s)"
		utils.ResponseHandling(req, w, msg)
		return
	}

	vars := mux.Vars(req)
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	newTodo, err := r.service.CreateTodo(ctx, input, *listId)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.AlreadyExistsErrorMsg) {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), utils.CreateErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to creat todo with name %s in list with id: %s", input.Name, listId)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Info(fmt.Sprintf("success creating todo with id: %s", newTodo.Id))
	utils.ResponseHandling(req, w, newTodo)
}

func (r *ResolverTodo) DeleteTodo(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	vars := mux.Vars(req)
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	todoId, err := utils.GetID(vars, todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	deletedTodo, err := r.service.DeleteTodo(ctx, *todoId, *listId)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), utils.DeletingErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to delete todo with id: %s", todoId)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success deleting todo with id: %s", todoId))
	utils.ResponseHandling(req, w, deletedTodo)
}

func (r *ResolverTodo) UpdateTodo(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	vars := mux.Vars(req)
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	todoId, err := utils.GetID(vars, todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	var input structures.TodoInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("error decoding todo body with id: %s", todoId)
		utils.ResponseHandling(req, w, msg)
		return
	}

	updatedTodo, err := r.service.UpdateTodo(ctx, *todoId, *listId, input)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.AlreadyExistsErrorMsg) {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), utils.UpdateErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to update todo with id: %s", todoId)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success updating todo with id: %s", updatedTodo.Id))
	utils.ResponseHandling(req, w, updatedTodo)
}

func (r *ResolverTodo) AssignUserToTodo(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	vars := mux.Vars(req)
	todoId, err := utils.GetID(vars, todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	username := req.Header.Get(username)
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		msg := "username is required"
		utils.ResponseHandling(req, w, msg)
		return
	}

	err = r.service.AssignUserToTodo(ctx, *todoId, *listId, username)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), assigningErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("error assigning user(%s) to task with id: %s", username, todoId)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("success assigning %s to todo with id: %s", username, todoId)
	utils.ResponseHandling(req, w, msg)
}

func (r *ResolverTodo) ChangeTodoStatus(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	vars := mux.Vars(req)
	todoId, err := utils.GetID(vars, todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	user := req.Header.Get(username)

	if utils.GetUsersRights(user) != utils.Role[utils.Admin] {
		todoAssignee := r.service.GetTodoAssignee(ctx, *todoId)

		if todoAssignee != user {
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("task %s is not assigned to %s", todoId, user)
			utils.ResponseHandling(req, w, msg)
			return
		}
	}

	err = r.service.ChangeTodoStatus(ctx, *todoId, *listId)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), changingStatusErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("error changing task status with id: %s", todoId)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("status successfuly changed to todo with id: %s", todoId)
	utils.ResponseHandling(req, w, msg)
}
