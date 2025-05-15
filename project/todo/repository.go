package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"project/structures"
	"project/utils"
	"strings"
	"time"
)

var (
	todoTable            = "todo"
	todoTableId          = "id"
	todoTableName        = "name"
	todoTableListId      = "list_id"
	todoTableStatus      = "status"
	todoTableAssignee    = "assignee"
	todoColumns          = []string{"id", "list_id", "name", "description", "deadline", "created_at", "assignee", "status", "priority"}
	insertTodoColumns    = []string{"id", "list_id", "name", "description", "deadline", "priority"}
	updateSetTodoColumns = []string{"name = ?", "description = ?", "deadline = ?", "priority = ?"}
	assignTodoColumn     = []string{"assignee = ?", "status = ?"}
)

type DBRepositoryTodo struct {
	db        *sqlx.DB
	converter RepositoryTodoConvertor
}

func NewDBRepositoryTodo(db *sqlx.DB, convertor RepositoryTodoConvertor) *DBRepositoryTodo {
	return &DBRepositoryTodo{db: db, converter: convertor}
}

func (r *DBRepositoryTodo) GetTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(todoColumns, ", "), todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var todoEntity structures.TodoEntity
	err := r.db.Get(&todoEntity, query, todoId, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting todo with id: %s", todoId))
		log.Error(err)
		return nil, err
	}

	todoModel := r.converter.ConvertEntityToModel(todoEntity)
	return &todoModel, nil
}

func (r *DBRepositoryTodo) GetAllTasks(ctx context.Context, listId uuid.UUID) []structures.TodoModel {
	cond := fmt.Sprintf(`%s = ?`, todoTableListId)
	sortBy := fmt.Sprintf(`ORDER BY %s`, todoTableName)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s %s`, strings.Join(todoColumns, ", "), todoTable, cond, sortBy)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var entities []structures.TodoEntity
	err := r.db.Select(&entities, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	return r.converter.ConvertEntitiesToModels(entities)
}

func (r *DBRepositoryTodo) CreateTodo(ctx context.Context, input structures.TodoEntity) error {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	stmt := fmt.Sprintf(`INSERT INTO %s(%s) VALUES(?, ?, ?, ?, ?, ?)`, todoTable, strings.Join(insertTodoColumns, ", "))
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, input.Id, input.ListId, input.Name, input.Description, input.Deadline, input.Priority)
	if err != nil {
		if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error already exists todo with the same name %s in list with id: %s", input.Name, input.ListId))
		} else if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", input.ListId))
		}

		log.Error(err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}
	if affected != 1 {
		err = errors.New(fmt.Sprintf("error creating todo with this name %s", input.Name))
		log.Error(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return err
}

func (r *DBRepositoryTodo) DeleteTodo(ctx context.Context, todoId, listId uuid.UUID) (*structures.TodoModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tx.Rollback()

	deletedTodo, err := r.GetTodo(ctx, todoId, listId)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE %s`, todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, todoId, listId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found todo with id %s in the list with id: %s", todoId, listId))
		}

		log.Error(err)
		return nil, err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if affectedRows != 1 {
		err = errors.New(fmt.Sprintf("error deleting todo with id: %s", todoId))
		log.Error(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return deletedTodo, err
}

func (r *DBRepositoryTodo) validate(originalTodo *structures.TodoEntity, updateTodo structures.TodoEntity) {
	if updateTodo.Name != "" {
		originalTodo.Name = updateTodo.Name
	}
	if updateTodo.Description != "" {
		originalTodo.Description = updateTodo.Description
	}
	var defaultTime time.Time
	if updateTodo.Deadline != defaultTime {
		originalTodo.Deadline = updateTodo.Deadline
	}
	if updateTodo.Priority != utils.Undefined && updateTodo.Priority != "" {
		originalTodo.Priority = updateTodo.Priority
	}
}

func (r *DBRepositoryTodo) UpdateTodo(ctx context.Context, updatedTask structures.TodoEntity, listId uuid.UUID) (*structures.TodoModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tx.Rollback()

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(todoColumns, ", "), todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var todoEntity structures.TodoEntity
	err = r.db.Get(&todoEntity, query, updatedTask.Id, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error not found todo with id: %s", updatedTask.Id))
		log.Error(err)
		return nil, err
	}
	r.validate(&todoEntity, updatedTask)

	cond = fmt.Sprintf(`%s = ?`, todoTableId)
	stmt = fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, todoTable, strings.Join(updateSetTodoColumns, ", "), cond)
	query = sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, todoEntity.Name, todoEntity.Description, todoEntity.Deadline, todoEntity.Priority, todoEntity.Id)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found todo with id %s in the list with id: %s", updatedTask.Id, listId))
		} else if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New("error todo with this name is already created")
		}

		log.Error(err)
		return nil, err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if affectedRows != 1 {
		err = errors.New(fmt.Sprintf("error updating todo with id: %s", todoId))
		log.Error(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	todoModel := r.converter.ConvertEntityToModel(todoEntity)
	return &todoModel, nil
}

func (r *DBRepositoryTodo) AssignTodoToUser(ctx context.Context, todoId, listId uuid.UUID, username string) error {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	if username == "" {
		err = errors.New(fmt.Sprintf("username is required"))
		log.Error(err)
		return err
	}

	assignee := r.GetTodoAssignee(ctx, todoId)
	if assignee != "" {
		err = errors.New(fmt.Sprintf("error assigning %s because %s is already assigned to todo with id: %s", username, assignee, todoId))
		log.Error(err)
		return err
	}

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, todoTable, strings.Join(assignTodoColumn, ", "), cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, username, utils.Assigned, todoId, listId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found todo with id %s in the list with id: %s", todoId, listId))
		}

		log.Error(err)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}
	if affectedRows != 1 {
		err = errors.New(fmt.Sprintf("error assigning %s to todo with id: %s", username, todoId))
		log.Error(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return err
}

func (r *DBRepositoryTodo) getStatus(todoId uuid.UUID) string {
	cond := fmt.Sprintf(`%s = ?`, todoTableId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, todoTableStatus, todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var status string
	err := r.db.Get(&status, query, todoId)
	if errors.Is(err, sql.ErrNoRows) {
		return utils.Undefined
	}

	return status
}

func (r *DBRepositoryTodo) ChangeTodoStatus(ctx context.Context, todoId, listId uuid.UUID) error {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`UPDATE %s SET %s = ? WHERE %s`, todoTable, todoTableStatus, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	currentStatus := r.getStatus(todoId)
	result, err := r.db.Exec(query, utils.NextStatus(currentStatus), todoId, listId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found todo with id %s in the list with id: %s", todoId, listId))
		}

		log.Error(err)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}
	if affectedRows != 1 {
		err = errors.New(fmt.Sprintf("error changing status to todo with id: %s", todoId))
		log.Error(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return err
}

func (r *DBRepositoryTodo) CheckIfListContainsTodo(ctx context.Context, todoId, listId uuid.UUID) bool {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, todoTableId, todoTableListId)
	stmt := fmt.Sprintf(`SELECT COUNT(%s) FROM %s WHERE %s`, todoTableId, todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var count int
	err := r.db.Get(&count, query, todoId, listId)
	if errors.Is(err, sql.ErrNoRows) {
		log.Error(err)
		return false
	}

	return count == 1
}

func (r *DBRepositoryTodo) GetTodoAssignee(ctx context.Context, todoId uuid.UUID) string {
	var assignee string
	cond := fmt.Sprintf(`%s = ?`, todoTableId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, todoTableAssignee, todoTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	err := r.db.Get(&assignee, query, todoId)
	if errors.Is(err, sql.ErrNoRows) {
		return ""
	}

	return assignee
}
