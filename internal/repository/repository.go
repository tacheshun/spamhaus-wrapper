package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"spamhaus-wrapper/graph/model"
	"spamhaus-wrapper/internal/service"
	"sync"
	"time"
)

const (
	fieldUUID         = "uuid"
	fieldIPAddress    = "ip_address"
	fieldResponseCode = "response_code"
	fieldCreatedAt    = "created_at"
	fieldUpdatedAt    = "updated_at"
)

type IPDetailsRepository struct {
	db                  *sql.DB
	mutex               sync.Mutex
	updateIPDetailsStmt *sql.Stmt
	getIPDetailsStmt    *sql.Stmt
}

func NewIPDetailsRepository(dbPath string) (*IPDetailsRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &IPDetailsRepository{db: db}, nil
}

func (r *IPDetailsRepository) getUpdateIPDetailsStmt() (*sql.Stmt, error) {
	if r.updateIPDetailsStmt == nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		if r.updateIPDetailsStmt == nil {
			stmt, err := r.db.Prepare(fmt.Sprintf(`
				INSERT INTO ip_details (%s, %s, %s, %s, %s)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(%s) DO UPDATE SET
					%s = excluded.%s,
					%s = excluded.%s
			`, fieldUUID, fieldIPAddress, fieldResponseCode, fieldCreatedAt, fieldUpdatedAt,
				fieldIPAddress,
				fieldResponseCode, fieldResponseCode,
				fieldUpdatedAt, fieldUpdatedAt))

			if err != nil {
				return nil, fmt.Errorf("failed to prepare updateIPDetails statement: %w", err)
			}
			r.updateIPDetailsStmt = stmt
		}
	}
	return r.updateIPDetailsStmt, nil
}

func (r *IPDetailsRepository) getGetIPDetailsStmt() (*sql.Stmt, error) {
	if r.getIPDetailsStmt == nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		if r.getIPDetailsStmt == nil {
			stmt, err := r.db.Prepare(fmt.Sprintf(`
				SELECT %s, %s, %s, %s, %s
				FROM ip_details
				WHERE %s = ?
			`, fieldUUID, fieldIPAddress, fieldResponseCode, fieldCreatedAt, fieldUpdatedAt, fieldIPAddress))

			if err != nil {
				return nil, fmt.Errorf("failed to prepare getIPDetails statement: %w", err)
			}
			r.getIPDetailsStmt = stmt
		}
	}
	return r.getIPDetailsStmt, nil
}

func (r *IPDetailsRepository) UpdateIPDetails(details *model.IPDetails) error {
	if details == nil {
		return fmt.Errorf("details cannot be nil")
	}

	if details.UUID == "" || details.IPAddress == "" || details.ResponseCode == "" {
		return fmt.Errorf("uuid, ip_address, and response_code must not be empty")
	}

	stmt, err := r.getUpdateIPDetailsStmt()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		details.UUID,
		details.IPAddress,
		details.ResponseCode,
		details.CreatedAt.Format(time.RFC3339),
		details.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to update IP details: %w", err)
	}

	return nil
}

func (r *IPDetailsRepository) UpdateMultipleIPDetails(ctx context.Context, ips []string) ([]*model.IPDetails, error) {
	spamhausService := service.NewSpamhausService()
	results := make([]*model.IPDetails, len(ips))
	var wg sync.WaitGroup

	for i, ip := range ips {
		wg.Add(1)
		go func(index int, ipAddress string) {
			defer wg.Done()

			responseCode, err := spamhausService.LookupIP(ctx, ipAddress)
			if err != nil {
				results[index] = &model.IPDetails{
					IPAddress:    ipAddress,
					ResponseCode: fmt.Sprintf("Error: %v", err),
				}
				return
			}

			details := &model.IPDetails{
				UUID:         generateUUID(),
				IPAddress:    ipAddress,
				ResponseCode: responseCode,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err = r.UpdateIPDetails(details)
			if err != nil {
				results[index] = &model.IPDetails{
					IPAddress:    ipAddress,
					ResponseCode: fmt.Sprintf("Error updating: %v", err),
				}
				return
			}

			results[index] = details
		}(i, ip)
	}

	wg.Wait()
	return results, nil
}

func (r *IPDetailsRepository) GetIPDetails(ipAddress string) (*model.IPDetails, error) {
	if ipAddress == "" {
		return nil, fmt.Errorf("ip_address must not be empty")
	}

	stmt, err := r.getGetIPDetailsStmt()
	if err != nil {
		return nil, err
	}

	var details model.IPDetails
	var createdAt, updatedAt string

	err = stmt.QueryRow(ipAddress).Scan(
		&details.UUID,
		&details.IPAddress,
		&details.ResponseCode,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No matching record found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get IP details: %w", err)
	}

	details.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	details.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &details, nil
}

func (r *IPDetailsRepository) Close() error {
	if r.updateIPDetailsStmt != nil {
		if err := r.updateIPDetailsStmt.Close(); err != nil {
			return fmt.Errorf("failed to close updateIPDetails statement: %w", err)
		}
	}
	if r.getIPDetailsStmt != nil {
		if err := r.getIPDetailsStmt.Close(); err != nil {
			return fmt.Errorf("failed to close getIPDetails statement: %w", err)
		}
	}
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func generateUUID() string {
	return uuid.New().String()
}
