package list

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"project/structures"
	"project/utils"
	"strings"
)

var (
	listTable               = "list"
	listTableId             = "id"
	usersListsTableListId   = "list_id"
	usersListsTableUsername = "username"
	usersListsTable         = "users_lists"
	listTableName           = "name"
	usersListsTableIsOwner  = "is_owner"
	usersListTableUsername  = "username"
	listColumns             = []string{"id", "name", "created_at"}
	usersListsColumns       = []string{"list_id", "username", "is_owner"}
	insertListColumn        = []string{"id", "name"}
	insertUsersListsColumn  = []string{"list_id", "username", "is_owner"}
)

type DBRepositoryList struct {
	db        *sqlx.DB
	convertor RepositoryConvertorList
}

func NewDBRepositoryList(db *sqlx.DB, convertor RepositoryConvertorList) *DBRepositoryList {
	return &DBRepositoryList{db: db, convertor: convertor}
}

func (r *DBRepositoryList) GetListById(ctx context.Context, listId uuid.UUID) (*structures.ListModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	cond := fmt.Sprintf(`%s = ?`, listTableId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(listColumns, ", "), listTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var listEntity structures.ListEntity
	err := r.db.Get(&listEntity, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting list by id: %s", listId))
		log.Error(err)
		return nil, err
	}

	owner, err := r.GetListOwner(ctx, listId)
	if err != nil {
		return nil, err
	}

	cond = fmt.Sprintf(`%s = ?`, usersListsTableListId)
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, usersListsTableUsername, usersListsTable, cond)
	query = sqlx.Rebind(sqlx.DOLLAR, stmt)
	var usernames []string
	err = r.db.Select(&usernames, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting owner of list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	return r.convertor.ConvertEntitiesToModel(&listEntity, usernames, owner.Username), nil
}

func (r *DBRepositoryList) GetAllLists(ctx context.Context) []*structures.ListModel {
	var listIds []uuid.UUID
	sortBy := fmt.Sprintf(`ORDER BY %s`, listTableName)
	stmt := fmt.Sprintf(`SELECT %s FROM %s %s`, listTableId, listTable, sortBy)
	err := r.db.Select(&listIds, stmt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	listModels := make([]*structures.ListModel, len(listIds))
	for i, id := range listIds {
		listModels[i], err = r.GetListById(ctx, id)
		if err != nil {
			return nil
		}
	}

	return listModels
}

func (r *DBRepositoryList) GetListOwner(ctx context.Context, listId uuid.UUID) (*structures.UserModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	cond := fmt.Sprintf(`%s = TRUE AND %s = ?`, usersListsTableIsOwner, usersListsTableListId)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(usersListsColumns, ", "), usersListsTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var userEntity structures.ListUserEntity
	err := r.db.Get(&userEntity, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting owner of list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	cond = fmt.Sprintf(`%s = ?`, listTableId)
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, listTableName, listTable, cond)
	query = sqlx.Rebind(sqlx.DOLLAR, stmt)
	var listName string
	err = r.db.Get(&listName, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err := errors.New(fmt.Sprintf("error getting name of list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	userModel := r.convertor.ConvertUserEntityToModel(userEntity, listName)
	return userModel, nil
}

func (r *DBRepositoryList) GetUserFromListById(ctx context.Context, listId uuid.UUID, username string) (*structures.UserModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, usersListsTableListId, usersListTableUsername)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(usersListsColumns, ", "), usersListsTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var userEntity structures.ListUserEntity
	err := r.db.Get(&userEntity, query, listId, username)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting user of list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	cond = fmt.Sprintf(`%s = ?`, listTableId)
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, listTableName, listTable, cond)
	query = sqlx.Rebind(sqlx.DOLLAR, stmt)
	var listName string
	err = r.db.Get(&listName, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New(fmt.Sprintf("error getting name of list with id: %s", listId))
		return nil, err
	}

	userModel := r.convertor.ConvertUserEntityToModel(userEntity, listName)
	return userModel, nil
}

func (r *DBRepositoryList) CreateList(ctx context.Context, entityList structures.ListEntity, entityUser structures.ListUserEntity) error {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	stmt := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (?, ?)`, listTable, strings.Join(insertListColumn, ", "))
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := tx.Exec(query, entityList.Id, entityList.Name)
	if err != nil {
		if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error already exists list with this name %s", entityList.Name))
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
		err = errors.New(fmt.Sprintf("error creating list with this name %s", entityList.Name))
		log.Error(err)
		return err
	}

	stmt = fmt.Sprintf(`INSERT INTO %s(%s) VALUES (?, ?, ?)`, usersListsTable, strings.Join(insertUsersListsColumn, ", "))
	query = sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err = tx.Exec(query, entityUser.ListId, entityUser.Username, entityUser.IsOwner)
	if err != nil {
		if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error already exists user with this name %s in list with id: %s", entityUser.Username, entityUser.ListId))
		} else if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", entityUser.ListId))
		}

		log.Error(err)
		return err
	}

	affectedRows, err = result.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}
	if affectedRows != 1 {
		err = errors.New(fmt.Sprintf("error creating user connection for %s with list with id: %s", entityUser.Username, entityUser.ListId))
		log.Error(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return err
}

func (r *DBRepositoryList) AddUserToList(ctx context.Context, entityUser structures.ListUserEntity) error {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return err
	}
	defer tx.Rollback()

	stmt := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (?, ?, ?)`, usersListsTable, strings.Join(insertUsersListsColumn, ", "))
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, entityUser.ListId, entityUser.Username, entityUser.IsOwner)
	if err != nil {
		if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error already exists user with this name %s in list with id: %s", entityUser.Username, entityUser.ListId))
		} else if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", entityUser.ListId))
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
		err = errors.New(fmt.Sprintf("error creating user connection for %s with list with id: %s", entityUser.Username, entityUser.ListId))
		log.Error(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return err
}

func (r *DBRepositoryList) DeleteList(ctx context.Context, listId uuid.UUID) (*structures.ListModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tx.Rollback()

	toBeDeleted, err := r.GetListById(ctx, listId)
	if err != nil {
		err = errors.New(fmt.Sprintf("error not found list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	cond := fmt.Sprintf(`%s = ?`, listTableId)
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE %s`, listTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, listId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", listId))
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
		err = errors.New(fmt.Sprintf("error deleting list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	return toBeDeleted, err
}

func (r *DBRepositoryList) RemoveUserUserFromList(ctx context.Context, entityUser structures.ListUserEntity) (*structures.UserModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	if entityUser.IsOwner {
		deletedList, err := r.DeleteList(ctx, entityUser.ListId)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		deletedOwner := structures.UserModel{
			ListId:   deletedList.Id,
			ListName: deletedList.Name,
			Username: entityUser.Username,
			IsOwner:  entityUser.IsOwner,
		}
		return &deletedOwner, nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tx.Rollback()

	removingFromList, err := r.GetListById(ctx, entityUser.ListId)
	if err != nil {
		err = errors.New(fmt.Sprintf("error not found list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	cond := fmt.Sprintf(`%s = ? AND %s = ?`, usersListTableUsername, usersListsTableListId)
	stmt := fmt.Sprintf(`DELETE FROM %s WHERE %s`, usersListsTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, entityUser.Username, entityUser.ListId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", entityUser.ListId))
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
		err = errors.New(fmt.Sprintf("error removing user with this name %s from list with id: %s",
			entityUser.Username, entityUser.ListId))
		log.Error(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	removedUser := structures.UserModel{
		ListId:   removingFromList.Id,
		ListName: removingFromList.Name,
		Username: entityUser.Username,
		IsOwner:  entityUser.IsOwner,
	}
	return &removedUser, nil
}

func (r *DBRepositoryList) UpdateList(ctx context.Context, listId uuid.UUID, newListName string) (*structures.ListModel, error) {
	log := ctx.Value(utils.Logger).(*logrus.Entry)

	if newListName == "" {
		err := errors.New("list name is required")
		log.Error(err)
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tx.Rollback()

	cond := fmt.Sprintf(`%s = ?`, listTableId)
	stmt := fmt.Sprintf(`UPDATE %s SET %s = ? WHERE %s`, listTable, listTableName, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	result, err := r.db.Exec(query, newListName, listId)
	if err != nil {
		if strings.Contains(err.Error(), utils.NotFoundSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error not found list with id: %s", listId))
		} else if strings.Contains(err.Error(), utils.AlreadyExistsSQLErrorMsg) {
			err = errors.New(fmt.Sprintf("error already exists list with this name %s", newListName))
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
		err = errors.New(fmt.Sprintf("error updating list with id: %s", listId))
		log.Error(err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	updatedList, err := r.GetListById(ctx, listId)
	return updatedList, err
}

func (r *DBRepositoryList) CheckIfListExists(ctx context.Context, listId uuid.UUID) bool {
	cond := fmt.Sprintf(`%s = ?`, listTableId)
	stmt := fmt.Sprintf(`SELECT COUNT(%s) FROM %s WHERE %s`, listTableId, listTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var count int
	err := r.db.Get(&count, query, listId)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return count == 1
}

func (r *DBRepositoryList) ContainsUserInList(ctx context.Context, listId uuid.UUID, username string) bool {
	cond := fmt.Sprintf(`%s = ? AND %s = ?`, usersListTableUsername, usersListsTableListId)
	stmt := fmt.Sprintf(`SELECT COUNT(%s) FROM %s WHERE %s`, usersListTableUsername, usersListsTable, cond)
	query := sqlx.Rebind(sqlx.DOLLAR, stmt)
	var count int
	err := r.db.Get(&count, query, username, listId)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return count == 1
}
