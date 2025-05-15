package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"project/graphql/graph"
	"project/graphql/graph/utils"
)

const (
	method    = "method"
	path      = "path"
	requestId = "requestId"
)

type GraphQLMiddleware struct {
	resolver *graph.Resolver
}

func NewGraphQLMiddleware(resolver *graph.Resolver) *GraphQLMiddleware {
	return &GraphQLMiddleware{
		resolver: resolver,
	}
}

func (gqlM *GraphQLMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logrus.WithContext(r.Context())

		log.Data = logrus.Fields{
			method:         r.Method,
			path:           r.URL.Path,
			utils.Username: r.Header.Get(utils.Username),
			requestId:      uuid.New().String(),
			utils.Status:   http.StatusText(http.StatusOK),
		}

		log.Info("Incoming GraphQL Request")
		ctx := context.WithValue(r.Context(), utils.Logger, log)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (gqlM *GraphQLMiddleware) SetUserInformationToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := r.Header.Get(utils.Username)
		role := utils.GetUserRole(username)

		if username == "" || role == utils.Unknown {
			logrus.Error("missing valuable information about the user")
			_, err := w.Write([]byte("missing valuable information about the user"))
			if err != nil {
				logrus.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, utils.Username, username)
		ctx = context.WithValue(ctx, utils.Role, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
