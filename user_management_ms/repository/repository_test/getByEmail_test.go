package repository_test_test

import (
	"testing"
	"user_management_ms/repository/query_repository"
	"user_management_ms/repository/repository_test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetByEmail_SQLMock(t *testing.T) {
	conn, mock := repository_test.SetupMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "email"}).
		AddRow(1, "test@example.com")

	// The email is passed as $1, and LIMIT is $2
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE email=\$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("test@example.com", 1).
		WillReturnRows(rows)

	repo := query_repository.NewUserQueryRepository()
	user, err := repo.GetUserByEmail(conn, "test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
