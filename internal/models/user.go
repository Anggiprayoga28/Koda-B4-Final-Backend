package models

import (
	"time"
)

type User struct {
	ID           int64      `json:"id" db:"id"`
	FullName     string     `json:"full_name" db:"full_name" validate:"required,min=2,max=100"`
	Email        string     `json:"email" db:"email" validate:"required,email"`
	Password     string     `json:"-" db:"password"`
	ProfileImage *string    `json:"profile_image,omitempty" db:"profile_image"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserRegisterRequest struct {
	FullName        string `json:"full_name" validate:"required,min=2,max=100"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID           int64     `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	ProfileImage *string   `json:"profile_image,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	FullName     string  `json:"full_name" validate:"required,min=2,max=100"`
	Email        string  `json:"email" validate:"required,email"`
	ProfileImage *string `json:"profile_image"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:           u.ID,
		FullName:     u.FullName,
		Email:        u.Email,
		ProfileImage: u.ProfileImage,
		CreatedAt:    u.CreatedAt,
	}
}
