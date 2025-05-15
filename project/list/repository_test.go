package list_test

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"project/list"
	"project/structures"
	"project/utils"
	"regexp"
	"testing"
	"time"
)

func TestRepositoryGetList(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		input       string
		mock        func()
		expected    string
		expectedErr error
	}{
		{
			name:  "getting existing list",
			input: utils.TestListName,
			mock: func() {
				rowsList := sqlxmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow(utils.TestListId, utils.TestListName, time.Now())
				mock.ExpectQuery(`SELECT id, name, created_at FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rowsList)
			},
			expected: utils.TestListName,
		}, {
			name:  "getting non-existing list",
			input: utils.TestListName,
			mock: func() {
				mock.ExpectQuery(`SELECT id, name, created_at FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnError(errors.New("list TestList does not exist"))
			},
			expectedErr: errors.New("list .+ does not exist"),
		}, {
			name:        "passing empty string name",
			input:       "",
			mock:        func() {},
			expectedErr: errors.New("list name is empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()
			actual, err := repo.GetListById(ctx, utils.TestListId)

			if err != nil {
				ok, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, ok)
				return
			}

			require.Equal(t, actual.Name, testCase.expected)
		})
	}
}

func TestRepositoryGetAll(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name     string
		mock     func()
		expected []string
	}{
		{
			name: "get all lists",
			mock: func() {
				rows := sqlxmock.NewRows([]string{"id"}).
					AddRow(uuid.UUID{1}).
					AddRow(uuid.UUID{2})
				mock.ExpectQuery(`SELECT id FROM list`).
					WillReturnRows(rows)

				rowsList := sqlxmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow(uuid.UUID{1}, utils.TestListName+"1", time.Now())
				mock.ExpectQuery(`SELECT id, name, created_at FROM list WHERE id = \$1`).
					WithArgs(uuid.UUID{1}).
					WillReturnRows(rowsList)

				rowsList = sqlxmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow(uuid.UUID{2}, utils.TestListName+"2", time.Now())
				mock.ExpectQuery(`SELECT id, name, created_at FROM list WHERE id = \$1`).
					WithArgs(uuid.UUID{2}).
					WillReturnRows(rowsList)
			},
			expected: []string{"TestList1", "TestList2"},
		}, {
			name: "empty lists",
			mock: func() {
				mock.ExpectQuery(`SELECT id FROM list`).
					WillReturnRows(sqlxmock.NewRows([]string{"id"}))
			},
			expected: []string(nil),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual := repo.GetAllLists(ctx)
			for i, expected := range testCase.expected {
				assert.Contains(t, actual[i].Name, expected)
			}
		})
	}
}

func TestRepositoryGetUser(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputListId uuid.UUID
		inputUserId string
		mock        func()
		expected    string
		expectedErr error
	}{
		{
			name:        "getting existing user",
			inputListId: utils.TestListId,
			inputUserId: utils.TestUsername,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"list_id", "username", "is_owner"}).
					AddRow(utils.TestListId, utils.TestUsername, true)
				mock.ExpectQuery(`SELECT list_id, username, is_owner FROM users_lists WHERE list_id = \$1 AND username = \$2`).
					WithArgs(utils.TestListId, utils.TestUsername).
					WillReturnRows(rows)

				rows = sqlxmock.NewRows([]string{"name"}).
					AddRow(utils.TestListName)
				mock.ExpectQuery(`SELECT name FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
			},
			expected: utils.TestUsername,
		}, {
			name:        "getting non-existing user",
			inputListId: utils.TestListId,
			inputUserId: utils.TestUsername,
			mock: func() {
				mock.ExpectQuery(`SELECT list_id, username, is_owner FROM users_lists WHERE list_id = \$1 AND username = \$2`).
					WithArgs(utils.TestListId, utils.TestUsername).
					WillReturnError(errors.New("user TestUser does not exist"))
			},
			expectedErr: errors.New("user .+ does not exist"),
		}, {
			name:        "passing empty string user id",
			inputListId: utils.TestListId,
			inputUserId: "",
			mock:        func() {},
			expectedErr: errors.New("username is empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual, err := repo.GetUserFromListById(ctx, testCase.inputListId, testCase.inputUserId)
			if err != nil {
				ok, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, ok)
				return
			}

			require.Equal(t, testCase.expected, actual.Username)
		})
	}
}

func TestRepositoryGetListOwner(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputListId uuid.UUID
		mock        func()
		expected    string
		expectedErr error
	}{
		{
			name:        "getting existing owner",
			inputListId: utils.TestListId,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"list_id", "username", "is_owner"}).
					AddRow(utils.TestListId, utils.TestUsername, true)
				mock.ExpectQuery(`SELECT list_id, username, is_owner FROM users_lists WHERE is_owner = TRUE AND list_id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
				rows = sqlxmock.NewRows([]string{"name"}).
					AddRow(utils.TestListName)
				mock.ExpectQuery(`SELECT name FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
			},
			expected: utils.TestUsername,
		}, {
			name:        "non existing list",
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectQuery(`SELECT list_id, username, is_owner FROM users_lists WHERE is_owner = TRUE AND list_id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(sqlxmock.NewRows([]string{"list_id", "username", "is_owner"}))
			},
			expectedErr: errors.New("error getting owner of list with id: +."),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual, err := repo.GetListOwner(ctx, testCase.inputListId)
			if err != nil {
				ok, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, ok)
				return
			}

			require.Equal(t, testCase.expected, actual.Username)
		})
	}
}

func TestRepositoryCreate(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name            string
		inputListEntity structures.ListEntity
		inputUserEntity structures.ListUserEntity
		mock            func()
		expected        error
	}{
		{
			name: "create new list",
			inputListEntity: structures.ListEntity{
				Id:   utils.TestListId,
				Name: utils.TestListName,
			},
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  true,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO list\(id, name\) VALUES \(\$1, \$2\)`).
					WithArgs(utils.TestListId, utils.TestListName).
					WillReturnResult(sqlxmock.NewResult(1, 1))

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users_lists\(list_id, username, is_owner\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(utils.TestListId, utils.TestUsername, true).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectCommit()
			},
			expected: nil,
		}, {
			name: "already existing list",
			inputListEntity: structures.ListEntity{
				Id:   utils.TestListId,
				Name: utils.TestListName,
			},
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  true,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO list\(id, name\) VALUES \(\$1, \$2\)`).
					WithArgs(utils.TestListId, utils.TestListName).
					WillReturnError(errors.New(utils.AlreadyExistsSQLErrorMsg))
			},
			expected: errors.New("error already exists list with this name .+"),
		}, {
			name: "created new list but not added to table",
			inputListEntity: structures.ListEntity{
				Id:   utils.TestListId,
				Name: utils.TestListName,
			},
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  true,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO list\(id, name\) VALUES \(\$1, \$2\)`).
					WithArgs(utils.TestListId, utils.TestListName).
					WillReturnResult(sqlxmock.NewResult(0, 0))
			},
			expected: errors.New("error creating list with this name .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual := repo.CreateList(ctx, testCase.inputListEntity, testCase.inputUserEntity)
			if actual != nil {
				ok, err := regexp.MatchString(testCase.expected.Error(), actual.Error())
				require.NoError(t, err)
				require.True(t, ok)
				return
			}

			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestRepositoryAddUser(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name            string
		inputUserEntity structures.ListUserEntity
		mock            func()
		expectedErr     error
	}{
		{
			name: "add new user",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users_lists\(list_id, username, is_owner\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(utils.TestListId, utils.TestUsername, false).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "already exists in this list",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users_lists\(list_id, username, is_owner\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(utils.TestListId, utils.TestUsername, false).
					WillReturnError(errors.New(utils.AlreadyExistsSQLErrorMsg))
			},
			expectedErr: errors.New("error already exists user with this name .+ in list with id: .+"),
		}, {
			name: "list does not exist",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users_lists\(list_id, username, is_owner\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(utils.TestListId, utils.TestUsername, false).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found list with id: .+"),
		}, {
			name: "add new user but not added to table",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users_lists\(list_id, username, is_owner\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(utils.TestListId, utils.TestUsername, false).
					WillReturnResult(sqlxmock.NewResult(0, 0))
			},
			expectedErr: errors.New("error creating user connection for .+ with list with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			err := repo.AddUserToList(ctx, testCase.inputUserEntity)
			if testCase.expectedErr != nil {
				ok, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, ok)
				return
			} else if testCase.expectedErr != nil {
				t.Error("Expected error but got nil")
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDelete(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputListId uuid.UUID
		mock        func()
		expectedErr error
	}{
		{
			name:        "delete existing list",
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:        "delete not existing list",
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found list with id: .+"),
		}, {
			name:        "list deleted but not from table",
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error deleting list with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			_, err := repo.DeleteList(ctx, testCase.inputListId)
			if err != nil {
				result, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, result)
				return
			} else if testCase.expectedErr != nil {
				t.Error("Expected error but got nil")
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryRemoveUser(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name            string
		inputUserEntity structures.ListUserEntity
		mock            func()
		expectedErr     error
	}{
		{
			name: "remove existing user (not list owner)",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM users_lists WHERE username = \$1 AND list_id = \$2`).
					WithArgs(utils.TestUsername, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "remove not existing user",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM users_lists WHERE username = \$1 AND list_id = \$2`).
					WithArgs(utils.TestUsername, utils.TestListId).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found list with id: .+"),
		}, {
			name: "remove existing user (list owner)",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  true,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "remove existing user but not removed form the table",
			inputUserEntity: structures.ListUserEntity{
				ListId:   utils.TestListId,
				Username: utils.TestUsername,
				IsOwner:  false,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM users_lists WHERE username = \$1 AND list_id = \$2`).
					WithArgs(utils.TestUsername, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error removing user with this name .+ from list with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			_, err := repo.RemoveUserUserFromList(ctx, testCase.inputUserEntity)
			if err != nil {
				result, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, result)
				return
			} else if testCase.expectedErr != nil {
				t.Error("Expected error but got nil")
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUpdate(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name         string
		inputListId  uuid.UUID
		inputNewName string
		mock         func()
		expectedErr  error
	}{
		{
			name:         "update existing list",
			inputListId:  utils.TestListId,
			inputNewName: utils.TestListName,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE list SET name = \$1 WHERE id = \$2`).
					WithArgs(utils.TestListName, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:         "update not existing list",
			inputListId:  utils.TestListId,
			inputNewName: utils.TestListName,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE list SET name = \$1 WHERE id = \$2`).
					WithArgs(utils.TestListName, utils.TestListId).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found list with id: .+"),
		}, {
			name:         "update with empty name",
			inputListId:  utils.TestListId,
			inputNewName: "",
			mock:         func() {},
			expectedErr:  errors.New("list name is required"),
		}, {
			name:         "update existing list but not changed in table",
			inputListId:  utils.TestListId,
			inputNewName: utils.TestListName,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE list SET name = \$1 WHERE id = \$2`).
					WithArgs(utils.TestListName, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error updating list with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			_, err := repo.UpdateList(ctx, testCase.inputListId, testCase.inputNewName)
			if err != nil {
				result, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, result)
				return
			} else if testCase.expectedErr != nil {
				t.Error("Expected error but got nil")
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryCheckIfListExists(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputListId uuid.UUID
		mock        func()
		expected    bool
	}{
		{
			name:        "contains existing list",
			inputListId: utils.TestListId,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"COUNT"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(id\) FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
			},
			expected: true,
		}, {
			name:        "does not contain list",
			inputListId: utils.TestListId,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"COUNT"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(id\) FROM list WHERE id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			ok := repo.CheckIfListExists(ctx, testCase.inputListId)
			require.Equal(t, testCase.expected, ok)
		})
	}
}

func TestRepositoryContainsUser(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := list.NewRepositoryListConvertor()
	repo := list.NewDBRepositoryList(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name          string
		inputListId   uuid.UUID
		inputUsername string
		mock          func()
		expected      bool
	}{
		{
			name:          "contains existing user",
			inputListId:   utils.TestListId,
			inputUsername: utils.TestUsername,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"COUNT"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(username\) FROM users_lists WHERE username = \$1 AND list_id = \$2`).
					WithArgs(utils.TestUsername, utils.TestListId).
					WillReturnRows(rows)
			},
			expected: true,
		}, {
			name:          "user is not contained in list",
			inputListId:   utils.TestListId,
			inputUsername: utils.TestUsername,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"COUNT"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(username\) FROM users_lists WHERE username = \$1 AND list_id = \$2`).
					WithArgs(utils.TestUsername, utils.TestListId).
					WillReturnRows(rows)
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			ok := repo.ContainsUserInList(ctx, testCase.inputListId, testCase.inputUsername)
			require.Equal(t, testCase.expected, ok)
		})
	}
}
