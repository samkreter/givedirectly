package datastore



import (
	"database/sql"
	"context"
	"time"
	"fmt"

	"github.com/pkg/errors"
	_ "github.com/lib/pq"

	"github.com/samkreter/givedirectly/types"
)

var (
	ErrNotFound = errors.New("not found")
)

type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new sqlStore for access postgres
func NewSQLStore(user, dbname, password, host string, port int) (*SQLStore, error){
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLStore{
		db: db,
	}, nil
}

// ListRequest returns all request from the database
func (s *SQLStore) ListRequest(ctx context.Context)  ([]*types.Request, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, email, title FROM requests")
	if err != nil {
		return nil, err
	}

	requests := []*types.Request{}

	defer rows.Close()
	for rows.Next() {
		request := &types.Request{}
		if err := rows.Scan(&request.ID, &request.Email, &request.Title); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// GetRequest returns the specific request
func (s *SQLStore) GetRequest(ctx context.Context, requestID int)  (*types.Request, error) {
	request := &types.Request{}

	row := s.db.QueryRowContext(ctx, "SELECT id, email, title FROM requests WHERE id=$1", requestID)
	if err := row.Scan(&request.ID, &request.Email, &request.Title); err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return request, nil
}

// DeleteRequest removes the request and updates the associated book to available
func (s *SQLStore) DeleteRequest(ctx context.Context, requestID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var title string
	row := tx.QueryRowContext(ctx, "SELECT title FROM requests WHERE id=$1", requestID)
	if err := row.Scan(&title); err != nil {
		tx.Rollback()
		switch {
		case err == sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	_, err = tx.ExecContext(ctx, "UPDATE books SET timeRequested='', available=true WHERE title=$1", title)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM requests WHERE id=$1", requestID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// CreateRequest creates a checks if a book is available. If it is, then it updates the book and
// creates a new request. Otherwise, it will return the book without creating the request. This is all
// handled within a transaction to make sure the book does not change availability while the func is running.
func (s *SQLStore) CreateRequest(ctx context.Context, request *types.Request) (*types.Book, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	book := &types.Book{}

	// Get associated book
	row := tx.QueryRowContext(ctx, "SELECT id, available, title, timeRequested FROM books WHERE title=$1", request.Title)
	if err := row.Scan(&book.ID, &book.Available, &book.Title, &book.TimeRequested); err != nil {
		tx.Rollback()
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
		return nil, err
	}

	// If the books not available, we rollback the transaction and return the book
	if !book.Available {
		tx.Rollback()
		return book, nil
	}

	// Update the book with the ISO-8601 formatted date/time
	book.TimeRequested = time.Now().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, "UPDATE books SET timeRequested=$1, available=false WHERE id=$2", book.TimeRequested, book.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO requests (email, title) VALUES ($1, $2)", request.Email, request.Title)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return book, nil
}

// EnsureDB ensures the db has the correct tables set up
func (s *SQLStore) EnsureDB() error {
	if err := s.createBookTable(); err != nil {
		return err
	}

	if err := s.createRequestTable(); err != nil {
		return err
	}

	return nil
}


func (s *SQLStore) createBookTable() error {
	const qry = `
		CREATE TABLE IF NOT EXISTS books (
			id serial PRIMARY KEY,
			available BOOLEAN NOT NULL,
			title TEXT NOT NULL,
			timeRequested TEXT NOT NULL DEFAULT ''
		)`

	if _, err := s.db.Exec(qry); err != nil {
		return errors.Errorf("failed to create book table withe error: %v", err)
	}

	return nil
}

func (s *SQLStore) createRequestTable() error {
	const qry = `
		CREATE TABLE IF NOT EXISTS requests (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			title TEXT NOT NULL
		)`

	if _, err := s.db.Exec(qry); err != nil {
		return errors.Errorf("failed to create requests table withe error: %v", err)
	}

	return nil
}


