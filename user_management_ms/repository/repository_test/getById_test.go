package repository_test_test

import (
	"testing"
	"user_management_ms/repository/query_repository"
	"user_management_ms/repository/repository_test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetByID_SQLMock(t *testing.T) {
	conn, mock := repository_test.SetupMockDB(t)
	rows := sqlmock.NewRows([]string{"id", "email"}).
		AddRow(1, "test@example.com")
	// Correct regex for GORM query
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1). // First arg is the ID, second is the LIMIT value
		WillReturnRows(rows)

	repo := query_repository.NewUserQueryRepository()
	user, err := repo.GetByID(conn, 1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
