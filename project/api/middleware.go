package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"project/utils"
)

const (
	username  = "userId"
	listId    = "listId"
	userId    = "userId"
	requestId = "requestId"
	path      = "path"
	method    = "method"
)

type AuthenticationMiddleware struct {
	resolver *ResolverList
}

func NewAuthenticationMiddleware(resolver *ResolverList) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		resolver: resolver,
	}
}

func (amw *AuthenticationMiddleware) getRole(r *http.Request) int {
	username := r.Header.Get(username)
	return utils.GetUsersRights(username)
}

func (amw *AuthenticationMiddleware) UserExistenceAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get(username)
		if utils.GetUsersRights(username) < utils.Role[utils.Reader] {
			log := r.Context().Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusUnauthorized).Warn(fmt.Sprintf("user %s does not exist", username))

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) CheckForReaderPermissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := amw.getRole(r)
		ctx := r.Context()

		username := r.Header.Get(username)
		listId, err := utils.ValidateStringID(mux.Vars(r)[listId])
		if err != nil {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusBadRequest).Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !(*amw.resolver).IsUserPartOfList(ctx, *listId, username) &&
			role != utils.Role[utils.Admin] {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusForbidden).Warn(fmt.Sprintf("%s is not authorized as reader in list: %s", username, listId))

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) CheckForWriterPermissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listIdLabel := listId
		username := r.Header.Get(username)
		listId := mux.Vars(r)[listId]
		role := amw.getRole(r)

		if role < utils.Role[utils.Writer] {
			log := r.Context().Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusForbidden).Warn(fmt.Sprintf("%s is not authorized as writer in list: %s", username, listId))

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), listIdLabel, listId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (amw *AuthenticationMiddleware) CheckForOwnerPermissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		username := r.Header.Get(username)
		listId, err := utils.ValidateStringID(mux.Vars(r)[listId])
		if err != nil {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusBadRequest).Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !(*amw.resolver).IsOwnerUserOwnerToListById(ctx, *listId, username) &&
			amw.getRole(r) != utils.Role[utils.Admin] {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusForbidden).Warn(fmt.Sprintf("%s is not owner nor admin to list: %s", username, listId))

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) CheckForAdminPermissions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		username := r.Header.Get(username)
		if utils.GetUsersRights(username) != utils.Role[utils.Admin] {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusForbidden).Warn(fmt.Sprintf("%s is not admin", username))

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (amw *AuthenticationMiddleware) CheckForUserExistenceInList(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		username := r.Header.Get(username)
		listId, err := utils.GetID(mux.Vars(r), listId)
		if err != nil {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithField(utils.Status, http.StatusBadRequest).Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !(*amw.resolver).IsUserPartOfList(ctx, *listId, username) &&
			amw.getRole(r) != utils.Role[utils.Admin] {
			log := ctx.Value(utils.Logger).(logrus.FieldLogger)
			log.WithFields(logrus.Fields{utils.Status: http.StatusForbidden})
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logrus.WithContext(r.Context())

		log.Data = logrus.Fields{
			method:       r.Method,
			path:         r.URL.Path,
			userId:       r.Header.Get(username),
			requestId:    uuid.New().String(),
			utils.Status: http.StatusText(http.StatusOK),
		}

		log.Info("Incoming request")
		ctx := context.WithValue(r.Context(), utils.Logger, log)
		ctx = context.WithValue(ctx, username, r.Header.Get(username))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
