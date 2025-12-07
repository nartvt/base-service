package entity

import "time"

// User represents the domain entity for a user.
// This is a pure domain model without any framework or database dependencies.
type User struct {
	ID           int64
	Username     string
	Email        string
	PhoneNumber  string
	FirstName    string
	LastName     string
	HashPassword string
	Avatar       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsDeleted checks if the user has been soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
