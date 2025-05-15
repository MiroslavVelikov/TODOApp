package utils

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	TestListName        = "TestList"
	TestUsername        = "TestUser"
	TestTodoName        = "TestTodo"
	TestTodoDescription = "TestDesc"
)

var (
	TestListId = uuid.UUID{1}
	TestTodoId = uuid.UUID{2}
)

func HelperGetContext() context.Context {
	ctx := context.Background()
	log := logrus.WithContext(ctx)

	log.Data = logrus.Fields{
		"method":    "Test",
		"path":      "Test",
		"userId":    TestUsername,
		"requestId": uuid.New().String(),
		"status":    http.StatusText(http.StatusOK),
	}

	ctx = context.WithValue(ctx, Logger, log)
	ctx = context.WithValue(ctx, "username", TestUsername)

	return ctx
}
