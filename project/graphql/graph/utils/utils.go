package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"project/graphql/graph/model"
	"time"
)

const (
	Username                  = "userId"
	Role                      = "role"
	Unknown                   = "unknown"
	Reader                    = "reader"
	Writer                    = "writer"
	Owner                     = "owner"
	Admin                     = "admin"
	BaseUrl                   = "http://localhost:8080"
	BasePath                  = "/todo/api"
	contentTypeKey            = "Content-Type"
	contentTypeValue          = "application/json"
	Logger                    = "logger"
	Status                    = "status"
	userDoesNotHavePermission = "user is %s and does not have %s permission"
	emptyRoleErrorMsg         = "providing role is required"
)

const (
	TestUsername = "TestUsername"
	TestListName = "TestListName"
	TestTodoName = "TestTodoName"
)

var (
	TestListId = uuid.UUID{1}
	TestTodoId = uuid.UUID{2}
)

var Client = &http.Client{Timeout: 10 * time.Second}

var RoleType = map[string]int{
	Unknown: -1,
	Reader:  1,
	Writer:  2,
	Owner:   3,
	Admin:   4,
}

var users = map[string]string{
	"Niki":  Admin,
	"Ivan":  Writer,
	"Miro":  Reader,
	"Yosif": Writer,
}

func GetUserRole(name string) string {
	rank, exists := users[name]
	if !exists {
		return Unknown
	}

	return rank
}

func CheckIfUserHasPermission(userRole string, permissionLevel string) bool {
	user, ok := RoleType[userRole]
	if !ok {
		return false
	}

	permission, ok := RoleType[permissionLevel]
	if !ok {
		return false
	}

	return permission <= user
}

type RequestSender struct{}

func NewRequestSender() *RequestSender {
	return &RequestSender{}
}

func (rs *RequestSender) SendRequest(requestType, route string, body any, headerData map[string]string, expectedStatus int) ([]byte, error, int) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	req, err := http.NewRequest(requestType, route, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err, http.StatusInternalServerError

	}

	req.Header.Set(contentTypeKey, contentTypeValue)
	for k, v := range headerData {
		req.Header.Set(k, v)
	}

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	if resp.StatusCode != expectedStatus {
		return nil, errors.New(string(result)), resp.StatusCode
	}

	return result, nil, resp.StatusCode
}

func GetTestingContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, Logger, logrus.NewEntry(logrus.StandardLogger()))

	return ctx
}

func GetListPosition(listId string, lists []*model.ListOutput) (int, error) {
	for pos, list := range lists {
		if list.ID == listId {
			return pos, nil
		}
	}

	return 0, errors.New("list not found")
}

func GetTodoPosition(todoId string, todos []*model.TodoOutput) (int, error) {
	for pos, todo := range todos {
		if todo.ID == todoId {
			return pos, nil
		}
	}

	return 0, errors.New("todo not found")
}

func ValidatePermission(ctx context.Context, permissionLevel string) error {
	role, ok := ctx.Value(Role).(string)
	if !ok || role == "" {
		return errors.New(emptyRoleErrorMsg)
	}
	if !CheckIfUserHasPermission(role, permissionLevel) {
		return fmt.Errorf(userDoesNotHavePermission, role, permissionLevel)
	}

	return nil
}
