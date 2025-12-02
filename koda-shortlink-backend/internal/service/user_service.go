package service

import (
	"errors"
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/repository"
	"koda-shortlink-backend/internal/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(userID int64) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user.ToResponse(), nil
}

func (s *UserService) UpdateProfile(userID int64, req *models.UpdateProfileRequest) error {
	if !utils.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Email != req.Email {
		exists, err := s.userRepo.EmailExists(req.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already in use")
		}
	}

	user.FullName = req.FullName
	user.Email = req.Email
	user.ProfileImage = req.ProfileImage

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update profile")
	}

	return nil
}

func (s *UserService) ChangePassword(userID int64, req *models.ChangePasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("passwords do not match")
	}

	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	valid, err := utils.VerifyPassword(req.CurrentPassword, user.Password)
	if err != nil || !valid {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(userID, hashedPassword); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}
