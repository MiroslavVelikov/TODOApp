package api

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/ast"
	"net/http"
	"project/graphql/graph"
	"project/graphql/graph/list"
	"project/graphql/graph/todo"
	"project/graphql/graph/utils"
)

func ServerHandler() {
	requestSender := utils.NewRequestSender()
	listConverter := list.NewListConverter()
	var listReqSender list.RequestSenderInterface = requestSender
	listService := list.NewServiceList(listConverter, &listReqSender)
	todoConverter := todo.NewTodoConverter()
	var todoReqSender todo.RequestSenderInterface = requestSender
	todoService := todo.NewServiceTodo(todoConverter, &todoReqSender)

	resolver := graph.NewResolver(listService, todoService)
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	gqlMiddleware := NewGraphQLMiddleware(resolver)

	router := mux.NewRouter()
	router.Use(gqlMiddleware.SetUserInformationToContext)
	router.Use(gqlMiddleware.LoggingMiddleware)
	router.Handle(utils.BasePath, srv)

	err := http.ListenAndServe(":8081", router)
	if err != nil {
		log.Fatal(err)
	}
}
