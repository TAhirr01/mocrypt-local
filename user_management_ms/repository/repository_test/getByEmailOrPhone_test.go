package repository_test_test

import (
	"testing"
	"user_management_ms/repository/query_repository"
	"user_management_ms/repository/repository_test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetByEmailOrPhone_SQLMock(t *testing.T) {
	conn, mock := repository_test.SetupMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "email", "phone"}).
		AddRow(1, "test@example.com", "0703735474")

	// Match GORM's actual query format
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."email"=\$1 OR "users"\."phone" =\$2 ORDER BY "users"\."id" LIMIT \$3`).
		WithArgs("test@example.com", "0703735474", 1).
		WillReturnRows(rows)

	repo := query_repository.NewUserQueryRepository()
	user, err := repo.GetUserByEmailOrPhone(conn, "test@example.com", "0703735474")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "0703735474", user.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}
