package todo_test

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"project/graphql/graph/model"
	"project/graphql/graph/todo"
	mocks "project/graphql/graph/todo/automock"
	"project/graphql/graph/utils"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputTodo           model.Todo
		inputRequestCreator string
		expected            model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully create todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					&model.Todo{
						Name: utils.TestTodoName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusCreated).
					Return([]byte("Returned new todo"), nil, http.StatusCreated).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned new todo")).
					Return(&model.TodoOutput{
						ID:   utils.TestTodoId.String(),
						Name: utils.TestTodoName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId: utils.TestListId.String(),
			inputTodo: model.Todo{
				Name: utils.TestTodoName,
			},
			inputRequestCreator: utils.TestUsername,
			expected: model.TodoOutput{
				ID:   utils.TestTodoId.String(),
				Name: utils.TestTodoName,
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					&model.Todo{
						Name: utils.TestTodoName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusCreated).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				return &mocks.ServiceConverterTodo{}
			},
			inputListId: utils.TestListId.String(),
			inputTodo: model.Todo{
				Name: utils.TestTodoName,
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to TodoOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPost, url,
					&model.Todo{
						Name: utils.TestTodoName,
					}, map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusCreated).
					Return([]byte("Returned new todo"), nil, http.StatusCreated).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned new todo")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId: utils.TestListId.String(),
			inputTodo: model.Todo{
				Name: utils.TestTodoName,
			},
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.CreateTodo(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputRequestCreator, &testCase.inputTodo)
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

func TestUpdateTodo(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", utils.TestListId, utils.TestTodoId)
	testNewName := utils.TestTodoName

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputTodoId         string
		inputRequestCreator string
		expected            model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully update todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				todoUpdate := model.UpdateTodoInput{
					Name: &testNewName,
				}
				reqSender.EXPECT().SendRequest(http.MethodPut, url, &todoUpdate,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned updated todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned updated todo")).
					Return(&model.TodoOutput{
						ID:   utils.TestTodoId.String(),
						Name: utils.TestTodoName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expected: model.TodoOutput{
				ID:   utils.TestTodoId.String(),
				Name: utils.TestTodoName,
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				todoUpdate := model.UpdateTodoInput{
					Name: &testNewName,
				}
				reqSender.EXPECT().SendRequest(http.MethodPut, url, &todoUpdate,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				return &mocks.ServiceConverterTodo{}
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to TodoOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				todoUpdate := model.UpdateTodoInput{
					Name: &testNewName,
				}
				reqSender.EXPECT().SendRequest(http.MethodPut, url, &todoUpdate,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned updated todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned updated todo")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			todoUpdate := model.UpdateTodoInput{
				Name: &testNewName,
			}

			actual, err := service.UpdateTodo(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputTodoId, testCase.inputRequestCreator, &todoUpdate)
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

func TestDeleteTodo(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", utils.TestListId, utils.TestTodoId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputTodoId         string
		inputRequestCreator string
		expected            model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully delete todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned deleted todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned deleted todo")).
					Return(&model.TodoOutput{
						ID:   utils.TestTodoId.String(),
						Name: utils.TestTodoName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expected: model.TodoOutput{
				ID:   utils.TestTodoId.String(),
				Name: utils.TestTodoName,
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				return &mocks.ServiceConverterTodo{}
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to TodoOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodDelete, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned deleted todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned deleted todo")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.DeleteTodo(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputTodoId, testCase.inputRequestCreator)
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

func TestAssignUserToTodo(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", utils.TestListId, utils.TestTodoId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		inputListId         string
		inputTodoId         string
		inputRequestCreator string
		expected            string
		expectedError       error
	}{
		{
			name: "successfully assign user to todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPatch, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned assign user to todo message"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expected:            "Returned assign user to todo message",
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPatch, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var converter todo.ServiceConverterTodo = todo.NewTodoConverter()
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.AssignUserToTodo(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputTodoId, testCase.inputRequestCreator)
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

func TestChangeTodoStatus(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s/status", utils.TestListId, utils.TestTodoId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		inputListId         string
		inputTodoId         string
		inputRequestCreator string
		expected            string
		expectedError       error
	}{
		{
			name: "successfully change todo status",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPatch, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned todo with changed status"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expected:            "Returned todo with changed status",
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodPatch, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var converter todo.ServiceConverterTodo = todo.NewTodoConverter()
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.ChangeTodoStatus(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputTodoId, testCase.inputRequestCreator)
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

func TestGetTodoFromList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todo/%s", utils.TestListId, utils.TestTodoId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputTodoId         string
		inputRequestCreator string
		expected            model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully get todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned requested todo")).
					Return(&model.TodoOutput{
						ID:   utils.TestTodoId.String(),
						Name: utils.TestTodoName,
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expected: model.TodoOutput{
				ID:   utils.TestTodoId.String(),
				Name: utils.TestTodoName,
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				return &mocks.ServiceConverterTodo{}
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to TodoOutput failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todo"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodoOutput([]byte("Returned requested todo")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputTodoId:         utils.TestTodoId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.GetTodoFromList(utils.GetTestingContext(), testCase.inputListId,
				testCase.inputTodoId, testCase.inputRequestCreator)
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

func TestGetTodosFromList(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todos", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputRequestCreator string
		expected            []*model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully get todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
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
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			expected: []*model.TodoOutput{
				&model.TodoOutput{
					ID:   utils.TestTodoId.String(),
					Name: utils.TestTodoName,
				},
			},
		}, {
			name: "sending request failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return(nil, errors.New("executing request have failed"), http.StatusBadRequest).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				return &mocks.ServiceConverterTodo{}
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("executing request have failed"),
		}, {
			name: "converting to TodosOutputs failed",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return(nil,
						errors.New("converting response failed")).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			expectedError:       errors.New("converting response failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			actual, err := service.GetTodosFromList(utils.GetTestingContext(),
				nil,
				nil,
				testCase.inputListId,
				testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Todos)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}

func TestGetTodosFromListPagination(t *testing.T) {
	url := fmt.Sprintf(utils.BaseUrl+utils.BasePath+"/list/%s/todos", utils.TestListId)

	testCases := []struct {
		name                string
		requestSender       func() *mocks.RequestSenderInterface
		converter           func() *mocks.ServiceConverterTodo
		inputListId         string
		inputRequestCreator string
		inputFirst          int32
		inputAfter          string
		expected            []*model.TodoOutput
		expectedError       error
	}{
		{
			name: "successfully list all todos in one page",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return([]*model.TodoOutput{
						&model.TodoOutput{
							ID:   uuid.UUID{1}.String(),
							Name: utils.TestTodoName + "1",
						}, &model.TodoOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestTodoName + "2",
						},
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			inputFirst:          2,
			expected: []*model.TodoOutput{
				&model.TodoOutput{
					ID:   uuid.UUID{1}.String(),
					Name: utils.TestTodoName + "1",
				}, &model.TodoOutput{
					ID:   uuid.UUID{2}.String(),
					Name: utils.TestTodoName + "2",
				},
			},
		}, {
			name: "successfully get only the first todo",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return([]*model.TodoOutput{
						&model.TodoOutput{
							ID:   uuid.UUID{1}.String(),
							Name: utils.TestTodoName + "1",
						}, &model.TodoOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestTodoName + "2",
						},
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			inputFirst:          1,
			expected: []*model.TodoOutput{
				&model.TodoOutput{
					ID:   uuid.UUID{1}.String(),
					Name: utils.TestTodoName + "1",
				},
			},
		}, {
			name: "successfully get all todos without the first one",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return([]*model.TodoOutput{
						&model.TodoOutput{
							ID:   uuid.UUID{1}.String(),
							Name: utils.TestTodoName + "1",
						}, &model.TodoOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestTodoName + "2",
						},
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			inputFirst:          100,
			inputAfter:          uuid.UUID{1}.String(),
			expected: []*model.TodoOutput{
				&model.TodoOutput{
					ID:   uuid.UUID{2}.String(),
					Name: utils.TestTodoName + "2",
				},
			},
		}, {
			name: "try to get out of range todos",
			requestSender: func() *mocks.RequestSenderInterface {
				reqSender := &mocks.RequestSenderInterface{}
				reqSender.EXPECT().SendRequest(http.MethodGet, url, nil,
					map[string]string{
						utils.Username: utils.TestUsername,
					}, http.StatusOK).
					Return([]byte("Returned requested todos"), nil, http.StatusOK).
					Once()

				return reqSender
			},
			converter: func() *mocks.ServiceConverterTodo {
				srvConverter := &mocks.ServiceConverterTodo{}
				srvConverter.EXPECT().ConvertResponseToTodosOutputs([]byte("Returned requested todos")).
					Return([]*model.TodoOutput{
						&model.TodoOutput{
							ID:   uuid.UUID{1}.String(),
							Name: utils.TestTodoName + "1",
						}, &model.TodoOutput{
							ID:   uuid.UUID{2}.String(),
							Name: utils.TestTodoName + "2",
						},
					}, nil).
					Once()

				return srvConverter
			},
			inputListId:         utils.TestListId.String(),
			inputRequestCreator: utils.TestUsername,
			inputAfter:          uuid.UUID{2}.String(),
			expectedError:       errors.New("todo is out of range"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			converterMock := testCase.converter()
			var converter todo.ServiceConverterTodo = converterMock
			reqSenderMock := testCase.requestSender()
			var reqSender todo.RequestSenderInterface = reqSenderMock
			service := todo.NewServiceTodo(converter, &reqSender)

			firstParam := &testCase.inputFirst
			var afterParam *string = nil
			if testCase.inputAfter != "" {
				afterParam = &testCase.inputAfter
			}

			actual, err := service.GetTodosFromList(utils.GetTestingContext(),
				firstParam,
				afterParam,
				testCase.inputListId,
				testCase.inputRequestCreator)
			if err != nil {
				require.Equal(t, testCase.expectedError, err)
				converterMock.AssertExpectations(t)
				reqSenderMock.AssertExpectations(t)
				return
			}

			require.Equal(t, testCase.expected, actual.Todos)
			converterMock.AssertExpectations(t)
			reqSenderMock.AssertExpectations(t)
		})
	}
}
