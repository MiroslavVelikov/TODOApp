package todo_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"project/structures"
	"project/todo"
	mocks "project/todo/automock"
	"project/utils"
	"regexp"
	"testing"
	"time"
)

func TestResolverGetTodo(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputTodoId    uuid.UUID
		inputListName  string
		expected       string
		expectedStatus int
	}{
		{
			name: "get existing task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodo(mock.Anything, utils.TestTodoId, utils.TestListId).Return(&structures.TodoOutput{
					Id:     utils.TestTodoId,
					Name:   utils.TestTodoName,
					ListId: utils.TestListId,
				}, nil).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			inputListName:  utils.TestListName,
			expected:       utils.TestTodoName,
			expectedStatus: http.StatusOK,
		}, {
			name: "get non-existing task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodo(mock.Anything, utils.TestTodoId, utils.TestListId).Return(nil,
					errors.New(fmt.Sprintf("error getting task with id: %s", utils.TestTodoId))).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			inputListName:  utils.TestListName,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo/api/%s/%s", testCase.inputListName, testCase.inputTodoId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": utils.TestListId.String(), "todoId": testCase.inputTodoId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.GetTodo(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
			if testCase.expectedStatus == http.StatusOK {
				result, err := regexp.MatchString(testCase.expected, rr.Body.String())
				require.NoError(t, err)
				require.True(t, result)
			}
		})
	}
}

func TestResolverGetAll(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputListId    uuid.UUID
		expected       []string
		expectedStatus int
	}{
		{
			name: "get all tasks",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetAllTasks(mock.Anything, utils.TestListId).Return([]structures.TodoOutput{
					{
						Id:     uuid.UUID{0},
						Name:   "TestTask0",
						ListId: utils.TestListId,
					}, {
						Id:     uuid.UUID{1},
						Name:   "TestTask1",
						ListId: utils.TestListId,
					}, {
						Id:     uuid.UUID{2},
						Name:   "TestTask2",
						ListId: utils.TestListId,
					},
				})
				return service
			},
			inputListId:    utils.TestListId,
			expected:       []string{"TestTask0", "TestTask1", "TestTask2"},
			expectedStatus: http.StatusOK,
		}, {
			name: "get from non-existing/empty list",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetAllTasks(mock.Anything, utils.TestListId).Return([]structures.TodoOutput(nil)).Once()
				return service
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo/api/%s", testCase.inputListId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.GetAllTasks(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
			for _, expected := range testCase.expected {
				result, err := regexp.MatchString(expected, rr.Body.String())
				require.NoError(t, err)
				require.True(t, result)
			}
		})
	}
}

func TestResolverCreate(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputTask      []byte
		inputListId    uuid.UUID
		expectedStatus int
	}{
		{
			name: "create new task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				tm, err := time.Parse(time.RFC3339, "2025-03-15T14:30:00Z")
				require.NoError(t, err)
				service.EXPECT().CreateTodo(mock.Anything,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
					},
					utils.TestListId).
					Return(&structures.TodoOutput{
						Id:          utils.TestTodoId,
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
						ListId:      utils.TestListId,
					},
						nil).
					Once()
				return service
			},
			inputListId: utils.TestListId,
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"deadline": "2025-03-15T14:30:00Z",
"priority": "%s"
}
`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			expectedStatus: http.StatusCreated,
		}, {
			name: "create new task with invalid data",
			service: func() *mocks.ServiceTodo {
				return nil
			},
			inputListId:    utils.TestListId,
			inputTask:      []byte(fmt.Sprintf(`{"name":"%s"}`, utils.TestTodoName)),
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "create new task that already exists",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				tm, err := time.Parse(time.RFC3339, "2025-03-15T14:30:00Z")
				require.NoError(t, err)
				service.EXPECT().CreateTodo(mock.Anything,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
					},
					utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error already exists todo with name %s in list with id: %s", utils.TestTodoName, utils.TestListId))).
					Once()
				return service
			},
			inputListId: utils.TestListId,
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"deadline": "2025-03-15T14:30:00Z",
"priority": "%s"
}`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			expectedStatus: http.StatusConflict,
		}, {
			name: "create new task in not existing list",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				tm, err := time.Parse(time.RFC3339, "2025-03-15T14:30:00Z")
				require.NoError(t, err)
				service.EXPECT().CreateTodo(mock.Anything,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
					},
					utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error not found list with id: %s", utils.TestListId))).
					Once()
				return service
			},
			inputListId: utils.TestListId,
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"deadline": "2025-03-15T14:30:00Z",
"priority": "%s"
}`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			expectedStatus: http.StatusNotFound,
		}, {
			name: "error creating task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				tm, err := time.Parse(time.RFC3339, "2025-03-15T14:30:00Z")
				require.NoError(t, err)
				service.EXPECT().CreateTodo(mock.Anything,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
					},
					utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error creating todo with name %s in list with id: %s", utils.TestTodoName, utils.TestListId))).
					Once()
				return service
			},
			inputListId: utils.TestListId,
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"deadline": "2025-03-15T14:30:00Z",
"priority": "%s"
}`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/todo/api/%s", testCase.inputListId), bytes.NewReader(testCase.inputTask))
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.CreateTodo(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverDelete(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputTodoId    uuid.UUID
		expectedStatus int
	}{
		{
			name: "delete todo",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().DeleteTodo(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(&structures.TodoOutput{
						Id:     utils.TestTodoId,
						ListId: utils.TestListId,
					}, nil).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusOK,
		}, {
			name: "delete todo with not existing list",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().DeleteTodo(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error not found todo with id: %s in list with id: %s", utils.TestTodoId, utils.TestListId))).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "delete todo but not from table",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().DeleteTodo(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error deleting todo with id: %s in list with id: %s", utils.TestTodoId, utils.TestListId))).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/todo/api/%s/%s", utils.TestListId, testCase.inputTodoId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": utils.TestListId.String(), "todoId": testCase.inputTodoId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.DeleteTodo(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputTask      []byte
		inputTodoId    uuid.UUID
		expectedStatus int
	}{
		{
			name: "update whole task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				tm, err := time.Parse(time.RFC3339, "2025-03-15T14:30:00Z")
				require.NoError(t, err)
				service.EXPECT().UpdateTodo(mock.Anything,
					utils.TestTodoId,
					utils.TestListId,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
					}).
					Return(&structures.TodoOutput{
						Id:          utils.TestTodoId,
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Deadline:    tm,
						Priority:    utils.MediumPriority,
						ListId:      utils.TestListId,
					}, nil).
					Once()

				return service
			},
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"deadline": "2025-03-15T14:30:00Z",
"priority": "%s"
}
`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusOK,
		}, {
			name: "update task partially",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().UpdateTodo(mock.Anything,
					utils.TestTodoId,
					utils.TestListId,
					structures.TodoInput{
						Name: utils.TestTodoName,
					}).
					Return(&structures.TodoOutput{
						Id:          utils.TestTodoId,
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Priority:    utils.MediumPriority,
						ListId:      utils.TestListId,
					}, nil).
					Once()

				return service
			},
			inputTask:      []byte(fmt.Sprintf(`{"name": "%s"}`, utils.TestTodoName)),
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusOK,
		}, {
			name: "update task with invalid data",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().UpdateTodo(mock.Anything,
					utils.TestTodoId,
					utils.TestListId,
					structures.TodoInput{}).
					Return(nil,
						errors.New("EOF")).
					Once()

				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusInternalServerError,
		}, {
			name: "update task that does not exist",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().UpdateTodo(mock.Anything,
					utils.TestTodoId,
					utils.TestListId,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Priority:    utils.MediumPriority,
					}).
					Return(nil,
						errors.New(fmt.Sprintf("error not found todo with id: %s", utils.TestTodoId))).
					Once()
				return service
			},
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"priority": "%s"
}
`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "update task that have the same name as existing task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().UpdateTodo(mock.Anything,
					utils.TestTodoId,
					utils.TestListId,
					structures.TodoInput{
						Name:        utils.TestTodoName,
						Description: utils.TestTodoDescription,
						Priority:    utils.MediumPriority,
					}).
					Return(nil,
						errors.New(fmt.Sprintf("error already exists todo with id: %s", utils.TestTodoId))).
					Once()
				return service
			},
			inputTask: []byte(fmt.Sprintf(`{
"name": "%s",
"description": "%s",
"priority": "%s"
}
`, utils.TestTodoName, utils.TestTodoDescription, utils.MediumPriority)),
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/todo/api/%s/%s", utils.TestListId, testCase.inputTodoId), bytes.NewReader(testCase.inputTask))
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": utils.TestListId.String(), "todoId": testCase.inputTodoId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.UpdateTodo(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverAssign(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		inputTodoId    uuid.UUID
		inputUsername  string
		expectedStatus int
	}{
		{
			name: "assign task",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().AssignUserToTodo(mock.Anything, utils.TestTodoId, utils.TestListId, utils.TestUsername).
					Return(nil).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusOK,
		}, {
			name: "empty username",
			service: func() *mocks.ServiceTodo {
				return nil
			},
			inputTodoId:    utils.TestTodoId,
			inputUsername:  "",
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "already assigned task or does not exists",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().AssignUserToTodo(mock.Anything, utils.TestTodoId, utils.TestListId, utils.TestUsername).
					Return(errors.New(fmt.Sprintf("error assigning %s to todo with id: %s", utils.TestUsername, utils.TestTodoId))).
					Once()

				return service
			},
			inputTodoId:    utils.TestTodoId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "not found list",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().AssignUserToTodo(mock.Anything, utils.TestTodoId, utils.TestListId, utils.TestUsername).
					Return(errors.New("error not found list with id: " + utils.TestListId.String())).
					Once()

				return service
			},
			inputTodoId:    utils.TestTodoId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/todo/api/%s/%s", utils.TestListName, testCase.inputTodoId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": utils.TestListId.String(), "todoId": testCase.inputTodoId.String()})
			req.Header.Set("userId", testCase.inputUsername)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.AssignUserToTodo(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverStatus(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceTodo
		ctx            context.Context
		inputTodoId    uuid.UUID
		expectedStatus int
	}{
		{
			name: "change todo status success",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodoAssignee(mock.Anything, utils.TestTodoId).
					Return(utils.TestUsername).
					Once()
				service.EXPECT().ChangeTodoStatus(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(nil).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusOK,
		}, {
			name: "todo does not have user assigned",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodoAssignee(mock.Anything, utils.TestTodoId).
					Return("").
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "todo has different user assigned",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodoAssignee(mock.Anything, utils.TestTodoId).
					Return("RandomUser").
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "change todo status not found",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodoAssignee(mock.Anything, utils.TestTodoId).
					Return(utils.TestUsername).
					Once()
				service.EXPECT().ChangeTodoStatus(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(errors.New(fmt.Sprintf("error not found todo with id %s in the list with id: %s", utils.TestTodoId, utils.TestListId))).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "change todo status failing table not changed",
			service: func() *mocks.ServiceTodo {
				service := &mocks.ServiceTodo{}
				service.EXPECT().GetTodoAssignee(mock.Anything, utils.TestTodoId).
					Return(utils.TestUsername).
					Once()
				service.EXPECT().ChangeTodoStatus(mock.Anything, utils.TestTodoId, utils.TestListId).
					Return(errors.New(fmt.Sprintf("error changing status to todo with id: %s", utils.TestTodoId))).
					Once()
				return service
			},
			inputTodoId:    utils.TestTodoId,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := todo.NewResolverWithService(testCase.service())

			req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/todo/api/%s/%s/status", utils.TestListId, testCase.inputTodoId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": utils.TestListId.String(), "todoId": testCase.inputTodoId.String()})
			req.Header.Set("userId", utils.TestUsername)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.ChangeTodoStatus(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}
