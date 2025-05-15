package list_test

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"project/graphql/graph/list"
	mocks "project/graphql/graph/list/automock"
	"project/graphql/graph/model"
	"project/graphql/graph/utils"
	"testing"
)

func TestCreateList(t *testing.T) {
	url := utils.BaseUrl + utils.BasePath + "/list"

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputList           model.List
		inputRequestCreator string
		expected            model.ListOutput
		expectedError       error
	}{
		{
			name: "successfully create list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					model.List{
						Name: utils.TestListName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusCreated).
					Return([]byte("Returned new list"), nil, http.StatusCreated).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned new list")).
					Return(&model.ListOutput{
						ID:   utils.TestListId.String(),
						Name: utils.TestListName,
					}, nil).
					Once()

				return srvConverter
			},
			inputList: model.List{
				Name: utils.TestListName,
			},
			inputRequestCreator: utils.TestUsername,
			expected: model.ListOutput{
				ID:   utils.TestListId.String(),
				Name: utils.TestListName,
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url, model.List{
					Name: utils.TestListName,
				}, map[string]string{
					utils.Username: utils.TestUsername,
				}, http.StatusCreated).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputList: model.List{
				Name: utils.TestListName,
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to ListOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url, model.List{
					Name: utils.TestListName,
				}, map[string]string{
					utils.Username: utils.TestUsername,
				}, http.StatusCreated).
					Return([]byte("Returned new list"), nil, http.StatusCreated).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned new list")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputList: model.List{
				Name: utils.TestListName,
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.CreateList(utils.GetTestingContext(), testCase.inputList, testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected.ID, actual.ID)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestAddUserToList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		inputListId         uuid.UUID
		inputRequestCreator string
		inputNewUser        model.User
		expected            string
		expectedError       error
	}{
		{
			name: "successfully added user to list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					model.User{
						Username: utils.TestUsername + "_new",
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("user added"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputNewUser: model.User{
				Username: utils.TestUsername + "_new",
			},
			expected: "user added",
		}, {
			name: "failed to add user to list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					model.User{
						Username: utils.TestUsername + "_new",
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputNewUser: model.User{
				Username: utils.TestUsername + "_new",
			},
			expectedError: errors.New("executing request have failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var converter list.ServiceConverterList = list.NewListConverter()
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.AddUserToList(utils.GetTestingContext(), testCase.inputListId.String(),
				testCase.inputRequestCreator, testCase.inputNewUser)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestUpdateListName(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputListId         uuid.UUID
		inputRequestCreator string
		inputListUpdate     model.List
		expected            string
		expectedError       error
	}{
		{
			name: "successfully updated list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPut, url,
					model.List{
						Name: utils.TestListName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("list updated"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("list updated")).
					Return(&model.ListOutput{
						ID:   utils.TestListId.String(),
						Name: utils.TestListName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputListUpdate: model.List{
				Name: utils.TestListName,
			},
			expected: utils.TestListName,
		}, {
			name: "failed to update list name",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPut, url,
					model.List{
						Name: utils.TestListName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputListUpdate: model.List{
				Name: utils.TestListName,
			},
			expectedError: errors.New("executing request have failed"),
		}, {
			name: "converting to ListOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPut, url, model.List{
					Name: utils.TestListName,
				}, map[string]string{
					utils.Username: utils.TestUsername,
				}, http.StatusOK).
					Return([]byte("Returned updated list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned updated list")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputListUpdate: model.List{
				Name: utils.TestListName,
			},
			expectedError: errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.UpdateListName(utils.GetTestingContext(), testCase.inputListId.String(),
				testCase.inputRequestCreator, testCase.inputListUpdate)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Name)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestDeleteList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputListId         uuid.UUID
		inputRequestCreator string
		expected            string
		expectedError       error
	}{
		{
			name: "successfully deleted list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("list deleted"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("list deleted")).
					Return(&model.ListOutput{
						ID:   utils.TestListId.String(),
						Name: utils.TestListName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expected:            utils.TestListName,
		}, {
			name: "failed to delete list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to ListOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned deleted list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned deleted list")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.DeleteList(utils.GetTestingContext(), testCase.inputListId.String(), testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Name)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestRemoveUserFromList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users/%s", utils.TestListId, utils.TestUsername)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputListId         uuid.UUID
		inputRequestCreator string
		inputRemoveUser     string
		expected            string
		expectedError       error
	}{
		{
			name: "successfully removed user from list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("removed user from list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToUserOutput([]byte("removed user from list")).
					Return(&model.UserOutput{
						ListID:   utils.TestListId.String(),
						ListName: utils.TestListName,
						Username: utils.TestUsername,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputRemoveUser:     utils.TestUsername,
			expected:            utils.TestUsername,
		}, {
			name: "failed to remove user from list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputRemoveUser:     utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to UserOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned removed user from list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToUserOutput([]byte("Returned removed user from list")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			inputRemoveUser:     utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.RemoveUserFromList(utils.GetTestingContext(), testCase.inputListId.String(),
				testCase.inputRemoveUser, testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Username)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputListId         uuid.UUID
		inputRequestCreator string
		expected            string
		expectedError       error
	}{
		{
			name: "successfully got list by id",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				todoUrl := url + "/todos"
				reqSender.EXPECT().SendRequest(http.MethodGet, todoUrl, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("returned list")).
					Return(&model.ListOutput{
						ID:   utils.TestListId.String(),
						Name: utils.TestListName,
					}, nil).
					Once()

				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return([]*model.TodoOutput{
						&model.TodoOutput{
							ID:   utils.TestTodoId.String(),
							Name: utils.TestTodoName,
						},
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expected:            utils.TestListName,
		}, {
			name: "failed to get list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to ListOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned requested list")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId,
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.GetList(utils.GetTestingContext(), testCase.inputListId.String(), testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Name)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetLists(t *testing.T) {
	url := utils.BaseUrl + utils.BasePath + "/list"

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputRequestCreator string
		expected            []string
		expectedError       error
	}{
		{
			name: "successfully got all lists",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("returned list")).
					Return([]*model.ListOutput{
						&model.ListOutput{
							Name: utils.TestListName + "1",
						}, &model.ListOutput{
							Name: utils.TestListName + "2",
						}}, nil).
					Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			expected: []string{
				utils.TestListName + "1",
				utils.TestListName + "2",
			},
		}, {
			name: "failed to get all lists",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to ListsOutputs failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested lists"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("Returned requested lists")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.GetLists(utils.GetTestingContext(), nil, nil, testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			actualOnlyNames := make([]string, len(actual.Lists))
			for i, list := range actual.Lists {
				actualOnlyNames[i] = list.Name
			}
			require.Equal(t, testCase.expected, actualOnlyNames)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetListsPagination(t *testing.T) {
	url := utils.BaseUrl + utils.BasePath + "/list"

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputRequestCreator string
		inputFirst          int32
		inputAfter          string
		expected            []string
		expectedError       error
	}{
		{
			name: "successfully got all lists",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("returned list")).
					Return([]*model.ListOutput{
						&model.ListOutput{
							Name: utils.TestListName + "1",
						}, &model.ListOutput{
							Name: utils.TestListName + "2",
						}}, nil).
					Once()

				return srvConverter
			},
			inputFirst:          int32(2),
			inputRequestCreator: utils.TestUsername,
			expected: []string{
				utils.TestListName + "1",
				utils.TestListName + "2",
			},
		}, {
			name: "successfully got only the first list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("returned list")).
					Return([]*model.ListOutput{
						&model.ListOutput{
							Name: utils.TestListName + "1",
						}, &model.ListOutput{
							Name: utils.TestListName + "2",
						}}, nil).
					Once()

				return srvConverter
			},
			inputFirst:          int32(1),
			inputRequestCreator: utils.TestUsername,
			expected: []string{
				utils.TestListName + "1",
			},
		}, {
			name: "skip the first list and get the other",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("returned list")).
					Return([]*model.ListOutput{
						&model.ListOutput{
							ID:   uuid.UUID{1}.String(),
							Name: utils.TestListName + "1",
						}, &model.ListOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestListName + "2",
						}}, nil).
					Once()

				return srvConverter
			},
			inputFirst:          int32(100),
			inputAfter:          uuid.UUID{1}.String(),
			inputRequestCreator: utils.TestUsername,
			expected: []string{
				utils.TestListName + "2",
			},
		}, {
			name: "try to get out of range lists",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned list"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListsOutputs([]byte("returned list")).
					Return([]*model.ListOutput{
						&model.ListOutput{
							Name: utils.TestListName + "1",
						}, &model.ListOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestListName + "2",
						}}, nil).
					Once()

				return srvConverter
			},
			inputAfter:          uuid.UUID{2}.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("lists is out of range"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			firstParam := &testCase.inputFirst
			var afterParam *string = nil
			if testCase.inputAfter != "" {
				afterParam = &testCase.inputAfter
			}

			actual, err := service.GetLists(utils.GetTestingContext(), firstParam, afterParam, testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			actualOnlyNames := make([]string, len(actual.Lists))
			for i, list := range actual.Lists {
				actualOnlyNames[i] = list.Name
			}
			require.Equal(t, testCase.expected, actualOnlyNames)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetUserFromList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users/%s", utils.TestListId, utils.TestUsername)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputRequestListId  string
		inputRequestedUser  string
		inputRequestCreator string
		expected            model.UserOutput
		expectedError       error
	}{
		{
			name: "successfully got requested user",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned user"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToUserOutput([]byte("returned user")).
					Return(&model.UserOutput{
						ListID:   utils.TestListId.String(),
						ListName: utils.TestListName,
						Username: utils.TestUsername,
					}, nil).Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestedUser:  utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expected: model.UserOutput{
				ListID:   utils.TestListId.String(),
				ListName: utils.TestListName,
				Username: utils.TestUsername,
			},
		}, {
			name: "failed to get the user",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestedUser:  utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to UserOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested user"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToUserOutput([]byte("Returned requested user")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestedUser:  utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.GetUserFromList(utils.GetTestingContext(), testCase.inputRequestListId,
				testCase.inputRequestedUser, testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, *actual)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetUsersFromList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/users", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterList
		inputRequestListId  string
		inputRequestCreator string
		expected            model.ListOutput
		expectedError       error
	}{
		{
			name: "successfully got all users from list",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("returned users"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("returned users")).
					Return(&model.ListOutput{
						ID:    utils.TestListId.String(),
						Name:  utils.TestListName,
						Owner: utils.TestUsername,
					}, nil).Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expected: model.ListOutput{
				ID:    utils.TestListId.String(),
				Name:  utils.TestListName,
				Owner: utils.TestUsername,
			},
		}, {
			name: "failed to get the users",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil,
						errors.New("executing request have failed"),
						http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				return &mocks.ServiceConverterList{}
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to ListOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested users"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterList {
				srvConverter := &mocks.ServiceConverterList{}
				srvConverter.EXPECT().ConvertResponseToListOutput([]byte("Returned requested users")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputRequestCreator: utils.TestUsername,
			inputRequestListId:  utils.TestListId.String(),
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter list.ServiceConverterList = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender list.RequestSenderInterface = reqSenderMock
			service := list.NewServiceList(converter, &reqSender)

			actual, err := service.GetUsersFromList(utils.GetTestingContext(),
				testCase.inputRequestListId,
				testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, *actual)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}
