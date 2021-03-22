package datastore



import (
	"database/sql"
	"context"
	"time"

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

func NewSQLStore(connStr string) (*SQLStore, error){
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &SQLStore{
		db: db,
	}, nil
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
	row := tx.QueryRowContext(ctx, "SELECT id, available, title, timeRequested FROM books WHERE title=?", request.Title)
	if err := row.Scan(&book.ID, &book.Available, book.Title, book.TimeRequested); err != nil {
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
	timeRequested := time.Now().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, "UPDATE books SET timeRequested=$1, available=false WHERE id=$2", timeRequested, book.ID)
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


func (s *SQLStore) createBookTable() error {
	const qry = `
		CREATE TABLE IF NOT EXISTS books (
			id serial PRIMARY KEY,
			available BOOLEAN NOT NULL,
			title text NOT NULL,
			timeRequested text,
		)`

	if _, err := s.db.Exec(qry); err != nil {
		return errors.Errorf("failed to create book table withe error: %v", err)
	}

	return nil
}

func (s *SQLStore) createRequestTable() error {
	const qry = `
		CREATE TABLE IF NOT EXISTS requests (
			id serial PRIMARY KEY,
			email text NOT NULL,
			title text NOT NULL,
		)`

	if _, err := s.db.Exec(qry); err != nil {
		return errors.Errorf("failed to create requests table withe error: %v", err)
	}

	return nil
}


