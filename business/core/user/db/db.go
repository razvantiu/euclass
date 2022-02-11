// Package db contains user related CRUD functionality.
package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user access.
type Store struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, usr User) error {

	return nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, usr User) error {

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, userID string) error {

	return nil
}

// Query retrieves a list of existing users from the database.
func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {

	return nil, nil
}
