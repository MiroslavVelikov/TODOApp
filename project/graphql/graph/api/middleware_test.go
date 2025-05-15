package api_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"project/graphql/graph/api"
	"project/graphql/graph/utils"
	"testing"
)

func TestSetUserInformationToContext(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		require.NoError(t, err)
	})

	testMiddleware := api.NewGraphQLMiddleware(nil)

	ctx := context.WithValue(context.Background(), utils.Logger, logrus.New())

	testCases := []struct {
		name     string
		headers  map[string]string
		expected []byte
	}{
		{
			name: "successfully set user information",
			headers: map[string]string{
				utils.Username: "Miro",
			},
			expected: []byte("Success"),
		}, {
			name: "fail to set user information, because of empty username",
			headers: map[string]string{
				utils.Username: "",
			},
			expected: []byte("missing valuable information about the user"),
		}, {
			name:     "fail to set user information, because of empty header",
			expected: []byte("missing valuable information about the user"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := testMiddleware.SetUserInformationToContext(testHandler)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)

			for key, value := range testCase.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req.WithContext(ctx))

			actual := rr.Body.Bytes()
			require.Equal(t, testCase.expected, actual)
			
		})
	}
}
