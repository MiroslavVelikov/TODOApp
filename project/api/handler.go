package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net/http"
	"project/list"
	"project/todo"
	"project/utils"
)

const (
	basePath = "/todo/api"
)

//go:generate mockery --name ResolverList --output=automock --with-expecter=true
type ResolverList interface {
	GetListById(w http.ResponseWriter, req *http.Request)
	GetAllLists(w http.ResponseWriter, req *http.Request)
	CreateList(w http.ResponseWriter, req *http.Request)
	DeleteList(w http.ResponseWriter, req *http.Request)
	UpdateList(w http.ResponseWriter, req *http.Request)
	AddUserToList(w http.ResponseWriter, req *http.Request)
	RemoveUserFromList(w http.ResponseWriter, req *http.Request)
	GetUserFromListById(w http.ResponseWriter, req *http.Request)
	GetUsersFromListById(w http.ResponseWriter, req *http.Request)
	IsOwnerUserOwnerToListById(ctx context.Context, listId uuid.UUID, username string) bool
	IsUserPartOfList(ctx context.Context, listId uuid.UUID, username string) bool
}

func ServerHandler() {
	db, err := utils.ConnectToDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	listRepoConvertor := list.NewRepositoryListConvertor()
	listRepository := list.NewDBRepositoryList(db, *listRepoConvertor)
	listSrvConvertor := list.NewServiceListConvertor()
	listService := list.NewServiceList(listRepository, *listSrvConvertor)
	listR := list.NewResolverList(listService)

	var lrInterface ResolverList = listR
	todoRepoConvertor := todo.NewRepositoryTodoConvertor()
	todoRepository := todo.NewDBRepositoryTodo(db, *todoRepoConvertor)
	todoServiceConvertor := todo.NewServiceTodoConvertor()
	todoService := todo.NewServiceTodo(todoRepository, *todoServiceConvertor)
	todoR := todo.NewResolverTodo(todoService)

	amw := NewAuthenticationMiddleware(&lrInterface)

	router := mux.NewRouter()
	router.Use(LoggingMiddleware)
	router.Use(amw.UserExistenceAuthentication)

	authenticationAdminSubrouter := router.PathPrefix(basePath + "/list").Subrouter()
	authenticationAdminSubrouter.Use(amw.CheckForAdminPermissions)
	authenticationAdminSubrouter.HandleFunc("", listR.GetAllLists).Methods(http.MethodGet)

	authenticationReaderSubrouter := router.PathPrefix(basePath).Subrouter()
	authenticationReaderSubrouter.Use(amw.CheckForReaderPermissions)
	authenticationReaderSubrouter.HandleFunc("/list/{listId}", listR.GetListById).Methods(http.MethodGet)

	authenticationWriterSubrouter := router.PathPrefix(basePath + "/list").Subrouter()
	authenticationWriterSubrouter.Use(amw.CheckForWriterPermissions)
	authenticationWriterSubrouter.HandleFunc("", listR.CreateList).Methods(http.MethodPost)

	authenticationForTodoAccessSubrouter := router.PathPrefix(basePath + "/list/{listId}").Subrouter()
	authenticationForTodoAccessSubrouter.Use(amw.CheckForUserExistenceInList)
	authenticationForTodoAccessSubrouter.HandleFunc("/todo/{todoId}", todoR.GetTodo).Methods(http.MethodGet)
	authenticationForTodoAccessSubrouter.HandleFunc("/todos", todoR.GetAllTasks).Methods(http.MethodGet)

	authenticationFroTodoModificationSubrouter := authenticationForTodoAccessSubrouter.PathPrefix("/todo").Subrouter()
	authenticationFroTodoModificationSubrouter.Use(amw.CheckForWriterPermissions)
	authenticationFroTodoModificationSubrouter.HandleFunc("", todoR.CreateTodo).Methods(http.MethodPost)
	authenticationFroTodoModificationSubrouter.HandleFunc("/{todoId}", todoR.UpdateTodo).Methods(http.MethodPut)
	authenticationFroTodoModificationSubrouter.HandleFunc("/{todoId}", todoR.DeleteTodo).Methods(http.MethodDelete)
	authenticationFroTodoModificationSubrouter.HandleFunc("/{todoId}", todoR.AssignUserToTodo).Methods(http.MethodPatch)
	authenticationFroTodoModificationSubrouter.HandleFunc("/{todoId}/status", todoR.ChangeTodoStatus).Methods(http.MethodPatch)

	authenticationOwnerSubrouter := router.PathPrefix(basePath + "/list/{listId}").Subrouter()
	authenticationOwnerSubrouter.Use(amw.CheckForOwnerPermissions)
	authenticationOwnerSubrouter.HandleFunc("", listR.UpdateList).Methods(http.MethodPut)
	authenticationOwnerSubrouter.HandleFunc("", listR.DeleteList).Methods(http.MethodDelete)
	authenticationOwnerSubrouter.HandleFunc("/users", listR.AddUserToList).Methods(http.MethodPost)
	authenticationOwnerSubrouter.HandleFunc("/users", listR.GetUsersFromListById).Methods(http.MethodGet)
	authenticationOwnerSubrouter.HandleFunc("/users/{userId}", listR.RemoveUserFromList).Methods(http.MethodDelete)
	authenticationOwnerSubrouter.HandleFunc("/users/{userId}", listR.GetUserFromListById).Methods(http.MethodGet)

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
