package repository

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"spamhaus-wrapper/graph/model"
	"testing"
	"time"
)

func TestIPDetailsRepository_GetIPDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &IPDetailsRepository{db: db}

	// Expect the prepare statement to be created once for all tests
	mock.ExpectPrepare("SELECT (.+) FROM ip_details")

	t.Run("Existing IP", func(t *testing.T) {
		expectedIP := &model.IPDetails{
			UUID:         "123e4567-e89b-12d3-a456-426614174000",
			IPAddress:    "192.168.1.1",
			ResponseCode: "127.0.0.2",
			CreatedAt:    func(t time.Time) *time.Time { return &t }(time.Now()),
			UpdatedAt:    func(t time.Time) *time.Time { return &t }(time.Now()),
		}

		rows := sqlmock.NewRows([]string{"uuid", "ip_address", "response_code", "created_at", "updated_at"}).
			AddRow(expectedIP.UUID, expectedIP.IPAddress, expectedIP.ResponseCode, expectedIP.CreatedAt, expectedIP.UpdatedAt)

		mock.ExpectQuery("SELECT (.+) FROM ip_details WHERE (.+)").
			WithArgs(expectedIP.IPAddress).
			WillReturnRows(rows)

		result, err := repo.GetIPDetails(expectedIP.IPAddress)

		assert.NoError(t, err)
		assert.Equal(t, expectedIP, result)
	})

	t.Run("Non-existing IP", func(t *testing.T) {
		nonExistingIP := "10.0.0.1"

		mock.ExpectQuery("SELECT (.+) FROM ip_details WHERE (.+)").
			WithArgs(nonExistingIP).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetIPDetails(nonExistingIP)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Empty IP", func(t *testing.T) {
		result, err := repo.GetIPDetails("")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "ip_address must not be empty", err.Error())
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM ip_details WHERE (.+)").
			WithArgs("192.168.1.1").
			WillReturnError(errors.New("database error"))

		result, err := repo.GetIPDetails("192.168.1.1")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get IP details")
	})

	// Check if all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIPDetailsRepository_UpdateIPDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &IPDetailsRepository{db: db}

	t.Run("Successful Update", func(t *testing.T) {
		details := &model.IPDetails{
			UUID:         "123e4567-e89b-12d3-a456-426614174000",
			IPAddress:    "192.168.1.1",
			ResponseCode: "127.0.0.2",
			CreatedAt:    func(t time.Time) *time.Time { return &t }(time.Now()),
			UpdatedAt:    func(t time.Time) *time.Time { return &t }(time.Now()),
		}

		mock.ExpectPrepare("INSERT INTO ip_details").
			ExpectExec().
			WithArgs(details.UUID, details.IPAddress, details.ResponseCode, details.CreatedAt, details.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateIPDetails(details)

		assert.NoError(t, err)
	})
}
