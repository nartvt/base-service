package mapper

import (
	"time"

	"base-service/internal/database/user"
	"base-service/internal/domain/entity"
)

// UserDBToEntity converts a database user model to a domain entity.
func UserDBToEntity(dbUser *user.User) *entity.User {
	if dbUser == nil {
		return nil
	}

	var avatar string
	if dbUser.Avatar.Valid {
		avatar = dbUser.Avatar.String
	}

	var deletedAt *time.Time
	if dbUser.DeletedAt.Valid {
		t := dbUser.DeletedAt.Time
		deletedAt = &t
	}

	return &entity.User{
		ID:           dbUser.ID,
		Username:     dbUser.Username,
		Email:        dbUser.Email,
		PhoneNumber:  dbUser.PhoneNumber,
		FirstName:    dbUser.FirstName,
		LastName:     dbUser.LastName,
		HashPassword: dbUser.HashPassword,
		Avatar:       avatar,
		CreatedAt:    dbUser.CreatedAt.Time,
		UpdatedAt:    dbUser.UpdatedAt.Time,
		DeletedAt:    deletedAt,
	}
}

// UserEntityToCreateParams converts a domain entity to database create params.
func UserEntityToCreateParams(entity *entity.User) *user.CreateUserParams {
	if entity == nil {
		return nil
	}

	return &user.CreateUserParams{
		Username:     entity.Username,
		Email:        entity.Email,
		PhoneNumber:  entity.PhoneNumber,
		FirstName:    entity.FirstName,
		LastName:     entity.LastName,
		HashPassword: entity.HashPassword,
	}
}
