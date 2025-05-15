package list_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"project/list"
	mocks "project/list/automock"
	"project/structures"
	"project/utils"
	"regexp"
	"testing"
)

func TestResolverGetListById(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputSearch    uuid.UUID
		output         string
		expectedStatus int
	}{
		{
			name: "get existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetListById(mock.Anything, utils.TestListId).Return(&structures.ListUserOutput{
					Id:    utils.TestListId,
					Name:  utils.TestListName,
					Owner: utils.TestUsername,
					Users: []string{utils.TestUsername},
				}, nil).Once()
				return srvMock
			},
			inputSearch:    utils.TestListId,
			output:         utils.TestListName,
			expectedStatus: http.StatusOK,
		}, {
			name: "try getting non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetListById(mock.Anything, utils.TestListId).Return(nil, errors.New("not existing list")).Once()
				return srvMock
			},
			inputSearch:    utils.TestListId,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo/api/%s", testCase.inputSearch.String()), nil)
			require.NoError(t, err)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputSearch.String()})

			rr := httptest.NewRecorder()

			resolver.GetListById(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
			if testCase.expectedStatus == http.StatusOK {
				result, err := regexp.MatchString(testCase.output, rr.Body.String())
				require.NoError(t, err)
				require.True(t, result)
			}
		})
	}
}

func TestResolverGetAllListNames(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		expectedStatus int
	}{
		{
			name: "get all lists",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetAllLists(mock.Anything).Return([]*structures.ListOutput{
					&structures.ListOutput{
						Name: utils.TestListName + "0",
					}, &structures.ListOutput{
						Name: utils.TestListName + "1",
					}, &structures.ListOutput{
						Name: utils.TestListName + "2",
					},
				})
				return srvMock
			},
			expectedStatus: http.StatusOK,
		}, {
			name: "get no lists",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetAllLists(mock.Anything).Return(nil)
				return srvMock
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodGet, "/todo/api/", nil)
			req = req.WithContext(utils.HelperGetContext())
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.GetAllLists(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverCreateList(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListName  []byte
		inputUsername  string
		expectedStatus int
	}{
		{
			name: "create new list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().CreateList(mock.Anything, mock.Anything, mock.Anything).
					Return(&structures.ListOutput{Id: utils.TestListId, Name: utils.TestListName, Owner: utils.TestUsername}, nil).
					Once()
				return srvMock
			},
			inputListName:  []byte(fmt.Sprintf(`{"listId": "%s"}`, utils.TestListName)),
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusCreated,
		}, {
			name: "create already existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().CreateList(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("error already exists list with this name "+utils.TestListName)).Once().
					Once()
				return srvMock
			},
			inputListName:  []byte(fmt.Sprintf(`{"listId": "%s"}`, utils.TestListName)),
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodPost, "/todo/api/", bytes.NewReader(testCase.inputListName))
			req = req.WithContext(utils.HelperGetContext())
			require.NoError(t, err)
			req.Header.Set("userId", testCase.inputUsername)

			rr := httptest.NewRecorder()

			resolver.CreateList(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverDeleteList(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		expectedStatus int
	}{
		{
			name: "delete existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().DeleteList(mock.Anything, utils.TestListId).
					Return(&structures.ListOutput{
						Id:    utils.TestListId,
						Name:  utils.TestListName,
						Owner: utils.TestUsername,
					}, nil).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "delete non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().DeleteList(mock.Anything, utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error not found list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "delete list but not from table",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().DeleteList(mock.Anything, utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("error deleting list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/todo/api/%s", testCase.inputListId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.DeleteList(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverUpdateList(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		inputNewList   []byte
		expectedStatus int
	}{
		{
			name: "update existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().UpdateList(mock.Anything, utils.TestListId, utils.TestListName).
					Return(&structures.ListOutput{
						Id:    utils.TestListId,
						Name:  utils.TestListName,
						Owner: utils.TestUsername,
					}, nil).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputNewList:   []byte(fmt.Sprintf(`{"name": "%s"}`, utils.TestListName)),
			expectedStatus: http.StatusOK,
		}, {
			name: "update non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().UpdateList(mock.Anything, utils.TestListId, utils.TestListName).
					Return(nil,
						errors.New(fmt.Sprintf("error not found list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputNewList:   []byte(fmt.Sprintf(`{"name": "%s"}`, utils.TestListName)),
			expectedStatus: http.StatusNotFound,
		}, {
			name: "update with invalid json body",
			service: func() *mocks.ServiceList {
				return nil
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "list with this name already exists",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().UpdateList(mock.Anything, utils.TestListId, utils.TestListName).
					Return(nil,
						errors.New(fmt.Sprintf("error already exists list with name %s", utils.TestListName))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputNewList:   []byte(fmt.Sprintf(`{"name": "%s"}`, utils.TestListName)),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodPut,
				fmt.Sprintf("/todo/api/%s", testCase.inputListId),
				bytes.NewReader(testCase.inputNewList))
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.UpdateList(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverAddUserToList(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		inputUsername  []byte
		expectedStatus int
	}{
		{
			name: "add user to existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().AddUserToList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(nil).Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  []byte(fmt.Sprintf(`{"username": "%s"}`, utils.TestUsername)),
			expectedStatus: http.StatusOK,
		}, {
			name: "add user to non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().AddUserToList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(errors.New(fmt.Sprintf("error not found list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputUsername:  []byte(fmt.Sprintf(`{"username": "%s"}`, utils.TestUsername)),
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "empty username",
			service: func() *mocks.ServiceList {
				return nil
			},
			inputUsername:  []byte(`{"username": ""}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name: "already added username",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().AddUserToList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(errors.New(fmt.Sprintf("error already exists user %s in list with id: %s", utils.TestUsername, utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  []byte(fmt.Sprintf(`{"username": "%s"}`, utils.TestUsername)),
			expectedStatus: http.StatusConflict,
		}, {
			name: "already added username",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().AddUserToList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(errors.New(fmt.Sprintf("error adding user %s in list with id: %s", utils.TestUsername, utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  []byte(fmt.Sprintf(`{"username": "%s"}`, utils.TestUsername)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodPost,
				fmt.Sprintf("/todo/api/%s", testCase.inputListId.String()),
				bytes.NewBuffer(testCase.inputUsername))
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.AddUserToList(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverRemoveUser(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		inputUsername  string
		expectedStatus int
	}{
		{
			name: "remove user from existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().RemoveUserFromList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(&structures.UserOutput{
						ListId:   utils.TestListId,
						ListName: utils.TestListName,
						Username: utils.TestUsername,
						IsOwner:  true,
					}, nil).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusOK,
		}, {
			name: "remove user from non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().RemoveUserFromList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(nil,
						errors.New(fmt.Sprintf("error not found list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusNotFound,
		}, {
			name: "remove non-existing username",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().RemoveUserFromList(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(nil,
						errors.New(fmt.Sprintf("error removing user %s from list with id: %s", utils.TestUsername, utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  utils.TestUsername,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/todo/api/%s/%s", testCase.inputListId, testCase.inputUsername), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String(), "userId": testCase.inputUsername})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.RemoveUserFromList(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestResolverGetUserFromListById(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		inputUsername  string
		expected       string
		expectedStatus int
	}{
		{
			name: "get user from existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetUserFromListById(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(&structures.UserOutput{
						ListId:   utils.TestListId,
						ListName: utils.TestListName,
						Username: utils.TestUsername,
						IsOwner:  false,
					}, nil).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  utils.TestUsername,
			expected:       utils.TestUsername,
			expectedStatus: http.StatusOK,
		}, {
			name: "get user from non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetUserFromListById(mock.Anything, utils.TestListId, utils.TestUsername).
					Return(nil,
						errors.New(fmt.Sprintf("failed to get user %s from list with id: %s", utils.TestUsername, utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			inputUsername:  utils.TestUsername,
			expected:       utils.TestUsername,
			expectedStatus: http.StatusInternalServerError,
		}, {
			name: "username is required",
			service: func() *mocks.ServiceList {
				return nil
			},
			inputListId:    utils.TestListId,
			inputUsername:  "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo/api/%s/%s",
				testCase.inputListId, testCase.inputUsername), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String(), "userId": testCase.inputUsername})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.GetUserFromListById(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
			result, err := regexp.MatchString(testCase.expected, rr.Body.String())
			require.NoError(t, err)
			require.True(t, result)
		})
	}
}

func TestResolverGetAllUsersFromListById(t *testing.T) {
	testCases := []struct {
		name           string
		service        func() *mocks.ServiceList
		inputListId    uuid.UUID
		expected       []string
		expectedStatus int
	}{
		{
			name: "get all users",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetUsersFromListById(mock.Anything, utils.TestListId).
					Return(&structures.ListUserOutput{
						Id:    utils.TestListId,
						Name:  utils.TestListName,
						Owner: utils.TestUsername,
						Users: []string{
							utils.TestUsername,
						},
					}, nil).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			expected:       []string{utils.TestUsername},
			expectedStatus: http.StatusOK,
		}, {
			name: "get users from non-existing list",
			service: func() *mocks.ServiceList {
				srvMock := &mocks.ServiceList{}
				srvMock.EXPECT().GetUsersFromListById(mock.Anything, utils.TestListId).
					Return(nil,
						errors.New(fmt.Sprintf("failed to get users from list with id: %s", utils.TestListId))).
					Once()
				return srvMock
			},
			inputListId:    utils.TestListId,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := list.NewResolverList(testCase.service())

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/todo/api/%s", testCase.inputListId), nil)
			req = req.WithContext(utils.HelperGetContext())
			req = mux.SetURLVars(req, map[string]string{"listId": testCase.inputListId.String()})
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			resolver.GetUsersFromListById(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
			for _, user := range testCase.expected {
				result, err := regexp.MatchString(user, rr.Body.String())
				require.NoError(t, err)
				require.True(t, result)
			}
		})
	}
}
