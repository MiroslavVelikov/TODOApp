package api_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"project/api"
	mocks "project/api/automock"
	"project/utils"
	"testing"
)

const (
	testList   = "TestList"
	testUser   = "TestUser"
	testWriter = "Yosif"
	userId     = "userId"
	listId     = "listId"
)

func TestUserExistenceAuthenticationMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "user exists",
			headers: map[string]string{
				userId: testWriter,
			},
			expectedStatus: http.StatusOK,
		}, {
			name: "user does not exist",
			headers: map[string]string{
				userId: testUser,
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = &mocks.ResolverList{}
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.UserExistenceAuthentication(testHandler)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req.WithContext(ctx))

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestCheckForReaderPermissions(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name           string
		resolver       func() *mocks.ResolverList
		headers        map[string]string
		inputParam     uuid.UUID
		expectedStatus int
	}{
		{
			name: "user is not admin but part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, "Miro").Return(true).Once()
				return resolver
			},
			headers: map[string]string{
				userId: "Miro",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is admin but not part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, "Niki").Return(false).Once()
				return resolver
			},
			headers: map[string]string{
				userId: "Niki",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is admin and part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, "Niki").Return(true).Once()
				return resolver
			},
			headers: map[string]string{
				userId: "Niki",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is not admin and nor part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, testUser).Return(false).Once()
				return resolver
			},
			headers: map[string]string{
				userId: testUser,
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = testCase.resolver()
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.CheckForReaderPermissions(testHandler)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testCase.inputParam), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{listId: testCase.inputParam.String()})
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestCheckForWriterPermissions(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	customReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testList), nil)
	customReq = mux.SetURLVars(customReq, map[string]string{"listName": testList})
	require.NoError(t, err)

	customReq.Header.Set(userId, testUser)

	testCases := []struct {
		name           string
		resolver       func() *mocks.ResolverList
		headers        map[string]string
		inputParam     uuid.UUID
		expectedStatus int
	}{
		{
			name: "user is writer or above writer",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				return resolver
			},
			headers: map[string]string{
				userId: "Ivan",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is reader",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				return resolver
			},
			headers: map[string]string{
				userId: "Miro",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		}, {
			name: "user is unknown",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				return resolver
			},
			headers: map[string]string{
				userId: "",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = testCase.resolver()
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.CheckForWriterPermissions(testHandler)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testCase.inputParam), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{listId: testCase.inputParam.String()})
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestCheckForOwnerPermissions(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name           string
		resolver       func() *mocks.ResolverList
		headers        map[string]string
		inputParam     uuid.UUID
		expectedStatus int
	}{
		{
			name: "user is owner",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsOwnerUserOwnerToListById(mock.Anything, utils.TestListId, testUser).Return(true).Once()
				return resolver
			},
			headers: map[string]string{
				userId: testUser,
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is admin",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsOwnerUserOwnerToListById(mock.Anything, utils.TestListId, "Niki").Return(false).Once()
				return resolver
			},
			headers: map[string]string{
				userId: "Niki",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is not owner nor admin",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsOwnerUserOwnerToListById(mock.Anything, utils.TestListId, testUser).Return(false).Once()
				return resolver
			},
			headers: map[string]string{
				userId: testUser,
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = testCase.resolver()
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.CheckForOwnerPermissions(testHandler)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testCase.inputParam), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{listId: testCase.inputParam.String()})
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestCheckForAdminPermissions(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name           string
		resolver       func() *mocks.ResolverList
		headers        map[string]string
		inputParam     uuid.UUID
		expectedStatus int
	}{
		{
			name: "user is admin",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				return resolver
			},
			headers: map[string]string{
				userId: "Niki",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is not admin",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				return resolver
			},
			headers: map[string]string{
				userId: "Miro",
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = testCase.resolver()
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.CheckForAdminPermissions(testHandler)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testCase.inputParam), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{listId: testCase.inputParam.String()})
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}

func TestCheckForUserExistenceInList(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name           string
		resolver       func() *mocks.ResolverList
		headers        map[string]string
		inputParam     uuid.UUID
		expectedStatus int
	}{
		{
			name: "user is part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, testUser).Return(true).Once()
				return resolver
			},
			headers: map[string]string{
				userId: testUser,
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusOK,
		}, {
			name: "user is not part of list",
			resolver: func() *mocks.ResolverList {
				resolver := &mocks.ResolverList{}
				resolver.EXPECT().IsUserPartOfList(mock.Anything, utils.TestListId, testUser).Return(false).Once()
				return resolver
			},
			headers: map[string]string{
				userId: testUser,
			},
			inputParam:     utils.TestListId,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var resolver api.ResolverList = testCase.resolver()
			middleware := api.NewAuthenticationMiddleware(&resolver)

			handler := middleware.CheckForUserExistenceInList(testHandler)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", testCase.inputParam), nil)
			req = req.WithContext(ctx)
			req = mux.SetURLVars(req, map[string]string{listId: testCase.inputParam.String()})
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}
