package datastore

import (
	"syreclabs.com/go/faker"

	"github.com/samkreter/givedirectly/types"
)

func(s *SQLStore) SeedDB(numToSeed int, testBooks... *types.Book) error {
	// Manually add books for easier testing
	for _, book := range testBooks {
		_, err := s.db.Exec("INSERT INTO books (available, title) VALUES ($1, $2)", book.Available, book.Title)
		if err != nil {
			return err
		}
	}

	// Add generated books
	for i:=0; i<numToSeed; i++ {
		title := faker.Lorem().Word()
		_, err := s.db.Exec("INSERT INTO books (available, title) VALUES ($1, $2)", true, title)
		if err != nil {
			return err
		}
	}

	return nil
}