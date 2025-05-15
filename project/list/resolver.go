package list

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
)

const (
	username           = "userId"
	listId             = "listId"
	removeUserErrorMsg = "error removing"
	addUserErrorMsg    = "error adding"
)

//go:generate mockery --name ServiceList --output=automock --with-expecter=true
type ServiceList interface {
	GetListById(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error)
	GetAllLists(ctx context.Context) []*structures.ListOutput
	GetUserFromListById(ctx context.Context, listId uuid.UUID, username string) (*structures.UserOutput, error)
	GetUsersFromListById(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error)
	CreateList(ctx context.Context, listName, username string) (*structures.ListOutput, error)
	AddUserToList(ctx context.Context, listId uuid.UUID, username string) error
	DeleteList(ctx context.Context, listId uuid.UUID) (*structures.ListUserOutput, error)
	RemoveUserFromList(ctx context.Context, listId uuid.UUID, username string) (*structures.UserOutput, error)
	UpdateList(ctx context.Context, listId uuid.UUID, newListName string) (*structures.ListOutput, error)
	CheckIfListExistsInList(ctx context.Context, listId uuid.UUID) bool
	ContainUserInList(ctx context.Context, listId uuid.UUID, username string) bool
}

type ResolverListImpl struct {
	service ServiceList
}

func NewResolverList(service ServiceList) *ResolverListImpl {
	return &ResolverListImpl{
		service: service,
	}
}

func (r *ResolverListImpl) getListIdInput(req *http.Request) (*uuid.UUID, error) {
	vars := mux.Vars(req)
	listId, err := utils.GetID(vars, listId)
	if err != nil {
		return nil, err
	}

	return listId, nil
}

func (r *ResolverListImpl) GetListById(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	list, err := r.service.GetListById(ctx, *listIdInput)
	if err != nil {
		isGetError, regErr := regexp.MatchString(utils.GetErrorMsg, err.Error())
		if isGetError && regErr == nil {
			w.WriteHeader(http.StatusNotFound)
			utils.ResponseHandling(req, w, err.Error())
			return
		}

		msg := fmt.Sprintf("error getting list with id: %s", *listIdInput)
		w.WriteHeader(http.StatusInternalServerError)
		utils.ResponseHandling(req, w, msg)
		return
	}

	log.Info(fmt.Sprintf("success getting %s list", list.Name))
	w.WriteHeader(http.StatusOK)
	utils.ResponseHandling(req.WithContext(ctx), w, list)
}

func (r *ResolverListImpl) GetAllLists(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	allLists := r.service.GetAllLists(ctx)

	w.WriteHeader(http.StatusOK)
	utils.ResponseHandling(req, w, allLists)
}

func (r *ResolverListImpl) CreateList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	var input structures.ListInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		msg := "error decoding list input"
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, msg)
		return
	}

	owner := req.Header.Get(username)
	newList, err := r.service.CreateList(ctx, input.Name, owner)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.AlreadyExistsErrorMsg) {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), utils.CreateErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to create list with name: %s", input.Name)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	log.WithField(utils.Status, http.StatusCreated).Info(fmt.Sprintf("success creating list with id: %s", newList.Id))
	w.WriteHeader(http.StatusCreated)
	utils.ResponseHandling(req, w, newList)
}

func (r *ResolverListImpl) DeleteList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	deletedList, err := r.service.DeleteList(ctx, *listIdInput)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), utils.DeletingErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
			utils.ResponseHandling(req, w, err.Error())
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to delete list with id: %s", listIdInput)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	logrus.Info(fmt.Sprintf("success deleting %s list", listIdInput))
	utils.ResponseHandling(req, w, deletedList)
}

func (r *ResolverListImpl) UpdateList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	var newVal structures.ListInput
	err = json.NewDecoder(req.Body).Decode(&newVal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("failed to decode new data for list with id: %s", listIdInput)
		utils.ResponseHandling(req, w, msg)
		return
	}

	updatedList, err := r.service.UpdateList(ctx, *listIdInput, newVal.Name)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), utils.AlreadyExistsErrorMsg) {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), utils.UpdateErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to decode new data for list with id: %s", listIdInput)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	logrus.Info(fmt.Sprintf("success updating %s list", listIdInput))
	utils.ResponseHandling(req, w, updatedList)
}

func (r *ResolverListImpl) AddUserToList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err)
		return
	}

	var userInput structures.ListUserInput
	err = json.NewDecoder(req.Body).Decode(&userInput)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := "failed to decode user"
		utils.ResponseHandling(req, w, msg)
		return
	}

	if userInput.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		msg := "user's username is required"
		utils.ResponseHandling(req, w, msg)
		return
	}

	err = r.service.AddUserToList(ctx, *listIdInput, userInput.Username)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.AlreadyExistsErrorMsg) {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), addUserErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to add user %s to list with id: %s", userInput.Username, listIdInput)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("success adding %s user to list with id: %s", userInput.Username, *listIdInput)
	utils.ResponseHandling(req, w, msg)
}

func (r *ResolverListImpl) RemoveUserFromList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	params := mux.Vars(req)
	username := params[username]

	removedUser, err := r.service.RemoveUserFromList(ctx, *listIdInput, username)
	if err != nil {
		msg := err.Error()
		if strings.Contains(err.Error(), utils.NotFoundErrorMsg) {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), removeUserErrorMsg) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			msg = fmt.Sprintf("failed to remove user %s from list with id: %s", username, listIdInput)
		}

		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success removing %s user from list with id: %s", username, *listIdInput))
	utils.ResponseHandling(req, w, removedUser)
}

func (r *ResolverListImpl) GetUserFromListById(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	params := mux.Vars(req)
	if params[username] == "" {
		w.WriteHeader(http.StatusBadRequest)
		msg := "user's username is required"
		utils.ResponseHandling(req, w, msg)
		return
	}

	username := params[username]
	user, err := r.service.GetUserFromListById(ctx, *listIdInput, username)
	if err != nil {
		isGetError, regErr := regexp.MatchString(utils.GetErrorMsg, err.Error())
		if isGetError && regErr == nil {
			w.WriteHeader(http.StatusNotFound)
			utils.ResponseHandling(req, w, err.Error())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("failed to get user %s from list with id: %s", username, listIdInput)
		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success getting %s user from list with id: %s", username, *listIdInput))
	utils.ResponseHandling(req, w, user)
}

func (r *ResolverListImpl) GetUsersFromListById(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	listIdInput, err := r.getListIdInput(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.ResponseHandling(req, w, err.Error())
		return
	}

	userOutputs, err := r.service.GetUsersFromListById(ctx, *listIdInput)
	if err != nil {
		isGetError, regErr := regexp.MatchString(utils.GetErrorMsg, err.Error())
		if isGetError && regErr == nil {
			w.WriteHeader(http.StatusNotFound)
			utils.ResponseHandling(req, w, err.Error())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("failed to get users from list with id: %s", listIdInput)
		utils.ResponseHandling(req, w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("success getting users from %s", listIdInput))
	utils.ResponseHandling(req, w, userOutputs)
}

func (r *ResolverListImpl) IsOwnerUserOwnerToListById(ctx context.Context, listId uuid.UUID, username string) bool {
	entity, err := r.service.GetUserFromListById(ctx, listId, username)
	if err != nil {
		return false
	}

	return entity.IsOwner
}

func (r *ResolverListImpl) IsUserPartOfList(ctx context.Context, listId uuid.UUID, username string) bool {
	return r.service.ContainUserInList(ctx, listId, username)
}
