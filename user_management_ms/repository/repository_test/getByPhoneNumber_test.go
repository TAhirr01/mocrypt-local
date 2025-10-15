package repository_test_test

import (
	"testing"
	"user_management_ms/repository/query_repository"
	"user_management_ms/repository/repository_test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetByPhoneNumber_SQLMock(t *testing.T) {
	conn, mock := repository_test.SetupMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "phone"}).
		AddRow(1, "0703735474")

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE phone=\$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("0703735474", 1).
		WillReturnRows(rows)

	repo := query_repository.NewUserQueryRepository()
	user, err := repo.GetUserByPhoneNumber(conn, "0703735474")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "0703735474", user.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}
