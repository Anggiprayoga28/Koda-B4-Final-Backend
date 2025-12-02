package models

import (
	"time"
)

type ShortLink struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	Destination string     `json:"destination" db:"destination"`
	UserID      *int64     `json:"user_id,omitempty" db:"user_id"`
	Title       *string    `json:"title,omitempty" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	ClickCount  int64      `json:"click_count" db:"click_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

type CreateLinkRequest struct {
	Destination string  `json:"destination" validate:"required,url"`
	CustomSlug  *string `json:"custom_slug,omitempty" validate:"omitempty,min=3,max=20,alphanum"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type UpdateLinkRequest struct {
	Destination *string `json:"destination,omitempty" validate:"omitempty,url"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type LinkResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	Destination string     `json:"destination"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type LinkListResponse struct {
	Links      []LinkResponse `json:"links"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

func (l *ShortLink) ToResponse(baseURL string) *LinkResponse {
	return &LinkResponse{
		ID:          l.ID,
		ShortCode:   l.ShortCode,
		ShortURL:    baseURL + "/" + l.ShortCode,
		Destination: l.Destination,
		Title:       l.Title,
		Description: l.Description,
		IsActive:    l.IsActive,
		ClickCount:  l.ClickCount,
		CreatedAt:   l.CreatedAt,
		ExpiresAt:   l.ExpiresAt,
	}
}
