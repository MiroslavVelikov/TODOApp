package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	postgres = "postgres"

	Unknown = "unknown"
	Reader  = "reader"
	Writer  = "writer"
	Owner   = "owner"
	Admin   = "admin"

	contentType     = "Content-Type"
	applicationJson = "application/json"

	Logger = "logger"
	Status = "status"

	AlreadyExistsErrorMsg    = "error already exists"
	NotFoundErrorMsg         = "not found"
	GetErrorMsg              = "error getting"
	CreateErrorMsg           = "error creating"
	DeletingErrorMsg         = "error deleting"
	UpdateErrorMsg           = "error updating"
	NotFoundSQLErrorMsg      = "violates foreign key constraint"
	AlreadyExistsSQLErrorMsg = "duplicate key value"
)

type Config struct {
	DbUser     string `envconfig:"DB_USER"`
	DbPassword string `envconfig:"DB_PWD"`
	DbName     string `envconfig:"DB_NAME"`
	DbHost     string `envconfig:"DB_HOST"`
	DbPort     string `envconfig:"DB_PORT"`
}

func testingPurposeFunc() Config {
	return Config{
		DbUser:     "postgres",
		DbPassword: "example",
		DbName:     "postgres",
		DbHost:     "localhost",
		DbPort:     "5433",
	}
}

func GetConnectionString() (string, error) {
	// Commented because of testing porous
	//var cfg Config
	//if err := envconfig.Init(&cfg); err != nil {
	//	return "", err
	//}
	cfg := testingPurposeFunc()

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName)
	fmt.Println(connectionString)
	return connectionString, nil
}

func ConnectToDB() (*sqlx.DB, error) {
	connectionString, err := GetConnectionString()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect(postgres, connectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetID(vars map[string]string, placeholderOfId string) (*uuid.UUID, error) {
	strId := vars[placeholderOfId]
	id, err := ValidateStringID(strId)

	return id, err
}

func ValidateStringID(idStr string) (*uuid.UUID, error) {
	if idStr == "" {
		return nil, errors.New("ID cannot be empty")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("invalid ID format, must be UUID")
	}

	return &id, nil
}

func ResponseHandling(request *http.Request, writer http.ResponseWriter, response any) {
	ctx := request.Context()
	log := ctx.Value(Logger).(*logrus.Entry)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("Error marshalling response")
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set(contentType, applicationJson)

	_, err = writer.Write(jsonResponse)
	if err != nil {
		log.WithError(err).Error("Error writing response")
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

var Role = map[string]int{
	Unknown: -1,
	Reader:  1,
	Writer:  2,
	Owner:   3,
	Admin:   4,
}

var users = map[string]int{
	"Niki":  Role[Admin],
	"Ivan":  Role[Writer],
	"Miro":  Role[Reader],
	"Yosif": Role[Writer],
}

func GetUsersRights(name string) int {
	rank, exists := users[name]
	if !exists {
		return -1
	}

	return rank
}
