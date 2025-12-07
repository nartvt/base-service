package biz

import (
	"context"
	"fmt"

	"base-service/internal/database/user"
	"base-service/internal/dto/request"
	"base-service/internal/dto/response"
	"base-service/internal/repository"
)

type UserBiz interface {
	ValidateUser(ctx context.Context, UsernameOrEmail string, hashPassword string) (*response.UserResponse, error)
	GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*user.User, error)
	RegisterUser(ctx context.Context, req *request.RegisterRequest) (response.ProfileResponse, error)
	GetUserProfile(ctx context.Context, userName string) (*response.ProfileResponse, error)
}

type userBizImpl struct {
	userRepository repository.UserRepository
}

func NewUserBiz(userRepository repository.UserRepository) UserBiz {
	return &userBizImpl{
		userRepository,
	}
}

func (b *userBizImpl) RegisterUser(ctx context.Context, req *request.RegisterRequest) (response.ProfileResponse, error) {
	userModel := user.CreateUserParams{
		Username:     req.UserName,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PhoneNumber:  req.Phone,
		Email:        req.Email,
		HashPassword: req.HashedPassword, // Save the hashed password
	}
	resp, err := b.userRepository.CreateUser(context.Background(), &userModel)
	if err != nil {
		return response.ProfileResponse{}, err
	}
	return response.ProfileResponse{
		Id:          resp.ID,
		Email:       resp.Email,
		DisplayName: fmt.Sprintf("%s %s", resp.FirstName, resp.LastName),
		Username:    resp.Username,
	}, nil
}

func (b *userBizImpl) GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*user.User, error) {
	return b.userRepository.GetUserByUsernameOrEmail(ctx, usernameOrEmail)
}

func (b *userBizImpl) GetUserProfile(ctx context.Context, username string) (*response.ProfileResponse, error) {
	user, err := b.userRepository.GetUserByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	return &response.ProfileResponse{
		Id:          user.ID,
		Email:       user.Email,
		DisplayName: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Username:    user.Username,
		CreatedAt:   user.CreatedAt.Time.UnixMilli(),
		UpdatedAt:   user.UpdatedAt.Time.UnixMilli(),
	}, nil
}

func (u *userBizImpl) ValidateUser(ctx context.Context, UsernameOrEmail string, hashPassword string) (*response.UserResponse, error) {
	req := user.ValidateUserPasswordByUserNameParams{
		Username:     UsernameOrEmail,
		HashPassword: hashPassword,
	}
	userModel, err := u.userRepository.ValidateUserPasswordByUserName(ctx, &req)
	if err != nil {
		return nil, err
	}
	if userModel.ID > 0 {
		return &response.UserResponse{
			Id:        userModel.ID,
			Username:  userModel.Username,
			Email:     userModel.Email,
			Phone:     userModel.PhoneNumber,
			FirstName: userModel.FirstName,
			LastName:  userModel.LastName,
			FullName:  fmt.Sprintf("%s %s", userModel.FirstName, userModel.LastName),
			CreatedAt: userModel.CreatedAt.Time.UnixMilli(),
			UpdatedAt: userModel.UpdatedAt.Time.UnixMilli(),
		}, nil
	}
	return nil, nil
}
