package todo_test

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"project/structures"
	"project/todo"
	"project/utils"
	"regexp"
	"testing"
	"time"
)

func TestRepositoryGetTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		mock        func()
		expected    string
		expectedErr error
	}{
		{
			name:        "getting existing todo",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)
			},
			expected: utils.TestTodoName,
		}, {
			name:        "getting non-existing todo",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errors.New("error getting todo with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual, err := repo.GetTodo(ctx, testCase.inputTodoId, utils.TestListId)
			if err != nil {
				result, err := regexp.MatchString(testCase.expectedErr.Error(), err.Error())
				require.NoError(t, err)
				require.True(t, result)
				return
			}

			require.Equal(t, testCase.expected, actual.Name)
		})
	}
}

func TestRepositoryGetAllTasks(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputListId uuid.UUID
		mock        func()
		expected    []string
	}{
		{
			name:        "get all lists",
			inputListId: utils.TestListId,
			mock: func() {
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(uuid.UUID{1}, utils.TestListId, "TestTask1", "TestDescription", time.Time{}, time.Time{},
						"TestUser", "assigned", "medium").
					AddRow(uuid.UUID{2}, utils.TestListId, "TestTask2", "TestDescription", time.Time{}, time.Time{},
						"TestUser", "assigned", "medium")
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority ` +
					`FROM todo WHERE list_id = \$1`).
					WithArgs(utils.TestListId).
					WillReturnRows(rows)
			},
			expected: []string{"TestTask1", "TestTask2"},
		}, {
			name:        "empty lists",
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority ` +
					`FROM todo WHERE list_id = \$2`).
					WithArgs(utils.TestListId).
					WillReturnRows(sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
						"created_at", "assignee", "status", "priority"}))
			},
			expected: make([]string, 0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actualEntities := repo.GetAllTasks(ctx, testCase.inputListId)
			actual := make([]string, len(actualEntities))
			for i, entity := range actualEntities {
				actual[i] = entity.Name
			}

			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestRepositoryCreateTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputEntity structures.TodoEntity
		mock        func()
		expectedErr error
	}{
		{
			name: "create new todo",
			inputEntity: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName,
				Description: utils.TestTodoDescription, Deadline: time.Time{}, Priority: utils.MediumPriority},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO todo\(id, list_id, name, description, deadline, priority\) VALUES\(\$1, \$2, \$3, \$4, \$5, \$6\)`).
					WithArgs(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name: "todo already existing in list",
			inputEntity: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName, Description: utils.TestTodoDescription,
				Deadline: time.Time{}, Priority: utils.MediumPriority, Status: utils.NotAssigned},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO todo\(id, list_id, name, description, deadline, priority\) VALUES\(\$1, \$2, \$3, \$4, \$5, \$6\)`).
					WithArgs(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority).
					WillReturnError(errors.New(utils.AlreadyExistsSQLErrorMsg))
			},
			expectedErr: errors.New("error already exists todo with the same name .+ in list with id: .+"),
		}, {
			name: "list is not found",
			inputEntity: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName, Description: utils.TestTodoDescription,
				Deadline: time.Time{}, Priority: utils.MediumPriority, Status: utils.NotAssigned},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO todo\(id, list_id, name, description, deadline, priority\) VALUES\(\$1, \$2, \$3, \$4, \$5, \$6\)`).
					WithArgs(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found list with id: .+"),
		}, {
			name: "created todo but not added in table",
			inputEntity: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName, Description: utils.TestTodoDescription,
				Deadline: time.Time{}, Priority: utils.MediumPriority, Status: utils.NotAssigned},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO todo\(id, list_id, name, description, deadline, priority\) VALUES\(\$1, \$2, \$3, \$4, \$5, \$6\)`).
					WithArgs(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority).
					WillReturnResult(sqlxmock.NewResult(0, 0))
			},
			expectedErr: errors.New("error creating todo with this name .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			err := repo.CreateTodo(ctx, testCase.inputEntity)
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

func TestRepositoryDeleteTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		mock        func()
		expectedErr error
	}{
		{
			name:        "delete existing todo",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)
				mock.ExpectExec(`DELETE FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:        "todo does not exist",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errors.New("error getting todo with id: .+"),
		}, {
			name:        "delete todo fails deleting it form table",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)
				mock.ExpectExec(`DELETE FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error deleting todo with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			_, err := repo.DeleteTodo(ctx, testCase.inputTodoId, utils.TestListId)
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

func TestRepositoryUpdateTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		inputUpdate structures.TodoEntity
		mock        func()
		expectedErr error
	}{
		{
			name:        "update existing todo",
			inputTodoId: utils.TestTodoId,
			inputUpdate: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName,
				Description: utils.TestTodoDescription, Deadline: time.Time{}, Priority: utils.MediumPriority},
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)

				mock.ExpectExec(`UPDATE todo SET name = \$1, description = \$2, deadline = \$3, priority = \$4 WHERE id = \$5`).
					WithArgs(utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority, utils.TestTodoId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:        "not existing todo",
			inputTodoId: utils.TestTodoId,
			inputUpdate: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName,
				Description: utils.TestTodoDescription, Deadline: time.Time{}, Priority: utils.MediumPriority},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errors.New("error not found todo with id: .+"),
		}, {
			name:        "update to already existing todo",
			inputTodoId: utils.TestTodoId,
			inputUpdate: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName,
				Description: utils.TestTodoDescription, Deadline: time.Time{}, Priority: utils.MediumPriority},
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)

				mock.ExpectExec(`UPDATE todo SET name = \$1, description = \$2, deadline = \$3, priority = \$4 WHERE id = \$5`).
					WithArgs(utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority, utils.TestTodoId).
					WillReturnError(errors.New(utils.AlreadyExistsSQLErrorMsg))
			},
			expectedErr: errors.New("error todo with this name is already created"),
		}, {
			name:        "update existing todo",
			inputTodoId: utils.TestTodoId,
			inputUpdate: structures.TodoEntity{Id: utils.TestTodoId, ListId: utils.TestListId, Name: utils.TestTodoName,
				Description: utils.TestTodoDescription, Deadline: time.Time{}, Priority: utils.MediumPriority},
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"id", "list_id", "name", "description", "deadline",
					"created_at", "assignee", "status", "priority"}).
					AddRow(utils.TestTodoId, utils.TestListId, utils.TestTodoName, utils.TestTodoDescription, time.Time{}, time.Time{},
						utils.TestUsername, utils.Assigned, utils.MediumPriority)
				mock.ExpectQuery(`SELECT id, list_id, name, description, deadline, created_at, assignee, status, priority `+
					`FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(rows)

				mock.ExpectExec(`UPDATE todo SET name = \$1, description = \$2, deadline = \$3, priority = \$4 WHERE id = \$5`).
					WithArgs(utils.TestTodoName, utils.TestTodoDescription, time.Time{}, utils.MediumPriority, utils.TestTodoId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error updating todo with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			_, err := repo.UpdateTodo(ctx, testCase.inputUpdate, utils.TestListId)
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

func TestRepositoryAssignUserToTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		inputUser   string
		mock        func()
		expectedErr error
	}{
		{
			name:        "assign user",
			inputTodoId: utils.TestTodoId,
			inputUser:   utils.TestUsername,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(`UPDATE todo SET assignee = \$1, status = \$2 WHERE id = \$3 AND list_id = \$4`).
					WithArgs(utils.TestUsername, utils.Assigned, utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:        "already assigned user",
			inputTodoId: utils.TestTodoId,
			inputUser:   utils.TestUsername,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnRows(
						sqlxmock.NewRows([]string{"assignee"}).
							AddRow(utils.TestUsername))
			},
			expectedErr: errors.New("error assigning .+ because .+ is already assigned to todo with id: .+"),
		}, {
			name:        "assign empty username",
			inputTodoId: utils.TestTodoId,
			inputUser:   "",
			mock: func() {
				mock.ExpectBegin()
			},
			expectedErr: errors.New("username is required"),
		}, {
			name:        "todo is not found",
			inputTodoId: utils.TestTodoId,
			inputUser:   utils.TestUsername,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(`UPDATE todo SET assignee = \$1, status = \$2 WHERE id = \$3 AND list_id = \$4`).
					WithArgs(utils.TestUsername, utils.Assigned, utils.TestTodoId, utils.TestListId).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found todo with id .+ in the list with id: .+"),
		}, {
			name:        "the update on assignee was not added in the table",
			inputTodoId: utils.TestTodoId,
			inputUser:   utils.TestUsername,
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(`UPDATE todo SET assignee = \$1, status = \$2 WHERE id = \$3 AND list_id = \$4`).
					WithArgs(utils.TestUsername, utils.Assigned, utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedErr: errors.New("error assigning .+ to todo with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			err := repo.AssignTodoToUser(ctx, testCase.inputTodoId, utils.TestListId, testCase.inputUser)
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

func TestRepositoryChangeStatusOfTodo(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		mock        func()
		expectedErr error
	}{
		{
			name:        "change status to existing task",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"status"}).
					AddRow(utils.InReview)
				mock.ExpectQuery(`SELECT status FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnRows(rows)
				mock.ExpectExec(`UPDATE todo SET status = \$1 WHERE id = \$2 AND list_id = \$3`).
					WithArgs(utils.Completed, utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		}, {
			name:        "change status to not existing task",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"status"}).
					AddRow(utils.InProgress)
				mock.ExpectQuery(`SELECT status FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnRows(rows)
				mock.ExpectExec(`UPDATE todo SET status = \$1 WHERE id = \$2 AND list_id = \$3`).
					WithArgs(utils.InReview, utils.TestTodoId, utils.TestListId).
					WillReturnError(errors.New(utils.NotFoundSQLErrorMsg))
			},
			expectedErr: errors.New("error not found todo with id .+ in the list with id: .+"),
		}, {
			name:        "the changed status was not saved in the table",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectBegin()
				rows := sqlxmock.NewRows([]string{"status"}).
					AddRow(utils.InProgress)
				mock.ExpectQuery(`SELECT status FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnRows(rows)
				mock.ExpectExec(`UPDATE todo SET status = \$1 WHERE id = \$2 AND list_id = \$3`).
					WithArgs(utils.InReview, utils.TestTodoId, utils.TestListId).
					WillReturnResult(sqlxmock.NewResult(0, 0))
			},
			expectedErr: errors.New("error changing status to todo with id: .+"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			err := repo.ChangeTodoStatus(ctx, testCase.inputTodoId, utils.TestListId)
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

func TestRepositoryContainsTodoInList(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		inputListId uuid.UUID
		mock        func()
		expected    bool
	}{
		{
			name:        "contains todo",
			inputTodoId: utils.TestTodoId,
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectQuery(`SELECT COUNT\(id\) FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(
						sqlxmock.NewRows([]string{"count"}).
							AddRow(1))
			},
			expected: true,
		}, {
			name:        "does not contains",
			inputTodoId: utils.TestTodoId,
			inputListId: utils.TestListId,
			mock: func() {
				mock.ExpectQuery(`SELECT COUNT\(id\) FROM todo WHERE id = \$1 AND list_id = \$2`).
					WithArgs(utils.TestTodoId, utils.TestListId).
					WillReturnRows(
						sqlxmock.NewRows([]string{"count"}).
							AddRow(0))
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual := repo.CheckIfListContainsTodo(ctx, testCase.inputTodoId, testCase.inputListId)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestRepositoryGetTodoAssignee(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	convertor := todo.NewRepositoryTodoConvertor()
	repo := todo.NewDBRepositoryTodo(db, *convertor)
	ctx := utils.HelperGetContext()

	testCases := []struct {
		name        string
		inputTodoId uuid.UUID
		mock        func()
		expected    string
	}{
		{
			name:        "get todo assignee",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$1`).
					WithArgs(utils.TestTodoId).
					WillReturnRows(
						sqlxmock.NewRows([]string{"assignee"}).
							AddRow(utils.TestUsername))
			},
			expected: utils.TestUsername,
		}, {
			name:        "todo does not exist",
			inputTodoId: utils.TestTodoId,
			mock: func() {
				mock.ExpectQuery(`SELECT assignee FROM todo WHERE id = \$2`).
					WithArgs(utils.TestTodoId).
					WillReturnError(sql.ErrNoRows)
			},
			expected: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mock()

			actual := repo.GetTodoAssignee(ctx, testCase.inputTodoId)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
