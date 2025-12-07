package mapper

import (
	"base-service/internal/adapter/http/dto/response"
	"base-service/internal/domain/entity"
)

// UserToProfileResponse converts a domain user entity to a profile response DTO.
func UserToProfileResponse(user *entity.User) *response.ProfileResponse {
	if user == nil {
		return nil
	}

	return &response.ProfileResponse{
		Id:          user.ID,
		Email:       user.Email,
		Active:      !user.IsDeleted(),
		DisplayName: user.FullName(),
		Avatar:      user.Avatar,
		Username:    user.Username,
		CreatedAt:   user.CreatedAt.UnixMilli(),
		UpdatedAt:   user.UpdatedAt.UnixMilli(),
	}
}

// UserToUserResponse converts a domain user entity to a user response DTO.
func UserToUserResponse(user *entity.User) *response.UserResponse {
	if user == nil {
		return nil
	}

	return &response.UserResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.PhoneNumber,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FullName:  user.FullName(),
		CreatedAt: user.CreatedAt.UnixMilli(),
		UpdatedAt: user.UpdatedAt.UnixMilli(),
	}
}
