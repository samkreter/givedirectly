package datastore

import (
	"syreclabs.com/go/faker"
)

func(s *SQLStore) SeedDB(numToSeed int) error {
	for i:=0; i<numToSeed; i++ {
		title := faker.Lorem().Sentence(2)
		_, err := s.db.Exec("INSERT INTO books (available, title) VALUES ($1, $2)", true, title)
		if err != nil {
			return err
		}
	}

	return nil
}