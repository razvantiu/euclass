package user

import (
	"time"
	"unsafe"

	"github.com/ardanlabs/service/business/core/user/db"
)

// User represents an individual user.
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	PasswordHash []byte    `json:"-"`
	DateCreated  time.Time `json:"date_created"`
	DateUpdated  time.Time `json:"date_updated"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// =============================================================================

func toUser(dbUsr db.User) User {
	pu := (*User)(unsafe.Pointer(&dbUsr))
	return *pu
}

func toUserSlice(dbUsrs []db.User) []User {
	users := make([]User, len(dbUsrs))
	for i, dbUsr := range dbUsrs {
		users[i] = toUser(dbUsr)
	}
	return users
}
