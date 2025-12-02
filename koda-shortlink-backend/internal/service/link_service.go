package service

import (
	"context"
	"errors"
	"fmt"
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/repository"
	"koda-shortlink-backend/internal/utils"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type LinkService struct {
	linkRepo    *repository.ShortLinkRepository
	clickRepo   *repository.ClickRepository
	redisClient *redis.Client
	baseURL     string
}

func NewLinkService(linkRepo *repository.ShortLinkRepository, clickRepo *repository.ClickRepository, redisClient *redis.Client, baseURL string) *LinkService {
	return &LinkService{
		linkRepo:    linkRepo,
		clickRepo:   clickRepo,
		redisClient: redisClient,
		baseURL:     baseURL,
	}
}

func (s *LinkService) CreateLink(req *models.CreateLinkRequest, userID *int64) (*models.LinkResponse, error) {
	destination, err := utils.NormalizeURL(req.Destination)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}

	var shortCode string
	if req.CustomSlug != nil && *req.CustomSlug != "" {
		if !utils.IsValidCustomSlug(*req.CustomSlug) {
			return nil, errors.New("invalid custom slug: must be 3-20 alphanumeric characters")
		}
		shortCode = utils.SanitizeCustomSlug(*req.CustomSlug)

		exists, _ := s.linkRepo.ShortCodeExists(shortCode)
		if exists {
			return nil, errors.New("custom slug already taken")
		}
	} else {
		for i := 0; i < 5; i++ {
			code, err := utils.GenerateShortCode(6)
			if err != nil {
				return nil, err
			}

			exists, _ := s.linkRepo.ShortCodeExists(code)
			if !exists {
				shortCode = code
				break
			}
		}
		if shortCode == "" {
			return nil, errors.New("failed to generate unique short code")
		}
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, errors.New("invalid expiry date format")
		}
		expiresAt = &parsed
	}

	link := &models.ShortLink{
		ShortCode:   shortCode,
		Destination: destination,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		IsActive:    true,
		ExpiresAt:   expiresAt,
	}

	if err := s.linkRepo.Create(link); err != nil {
		return nil, errors.New("failed to create link")
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("link:%s:destination", shortCode)
	s.redisClient.Set(ctx, cacheKey, destination, time.Hour)

	return link.ToResponse(s.baseURL), nil
}

func (s *LinkService) GetLinkByShortCode(shortCode string, userID int64) (*models.LinkResponse, error) {
	link, err := s.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, errors.New("link not found")
	}

	if link.UserID != nil && *link.UserID != userID {
		return nil, errors.New("unauthorized access to this link")
	}

	return link.ToResponse(s.baseURL), nil
}

func (s *LinkService) GetUserLinks(userID int64, page, pageSize int) (*models.LinkListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	links, total, err := s.linkRepo.FindByUser(userID, page, pageSize)
	if err != nil {
		return nil, errors.New("failed to retrieve links")
	}

	linkResponses := make([]models.LinkResponse, len(links))
	for i, link := range links {
		linkResponses[i] = *link.ToResponse(s.baseURL)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.LinkListResponse{
		Links:      linkResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *LinkService) UpdateLink(shortCode string, userID int64, req *models.UpdateLinkRequest) error {
	link, err := s.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return errors.New("link not found")
	}

	if link.UserID == nil || *link.UserID != userID {
		return errors.New("unauthorized to update this link")
	}

	if req.Destination != nil {
		destination, err := utils.NormalizeURL(*req.Destination)
		if err != nil {
			return errors.New("invalid URL format")
		}
		link.Destination = destination

		ctx := context.Background()
		cacheKey := fmt.Sprintf("link:%s:destination", shortCode)
		s.redisClient.Del(ctx, cacheKey)
	}

	if req.Title != nil {
		link.Title = req.Title
	}

	if req.Description != nil {
		link.Description = req.Description
	}

	if req.IsActive != nil {
		link.IsActive = *req.IsActive
	}

	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return errors.New("invalid expiry date format")
		}
		link.ExpiresAt = &parsed
	}

	if err := s.linkRepo.Update(link); err != nil {
		return errors.New("failed to update link")
	}

	return nil
}

func (s *LinkService) DeleteLink(shortCode string, userID int64) error {
	link, err := s.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return errors.New("link not found")
	}

	if link.UserID == nil || *link.UserID != userID {
		return errors.New("unauthorized to delete this link")
	}

	if err := s.linkRepo.Delete(link.ID, userID); err != nil {
		return errors.New("failed to delete link")
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("link:%s:destination", shortCode)
	s.redisClient.Del(ctx, cacheKey)

	return nil
}

func (s *LinkService) GetDestination(shortCode string) (string, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("link:%s:destination", shortCode)

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		return cached, nil
	}

	link, err := s.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return "", errors.New("link not found")
	}

	if !link.IsActive {
		return "", errors.New("link is inactive")
	}

	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return "", errors.New("link has expired")
	}

	s.redisClient.Set(ctx, cacheKey, link.Destination, time.Hour)

	return link.Destination, nil
}

func (s *LinkService) RecordClick(shortCode string, click *models.Click) error {
	link, err := s.linkRepo.FindByShortCode(shortCode)
	if err != nil {
		return err
	}

	click.LinkID = link.ID

	if err := s.clickRepo.Create(click); err != nil {
		return err
	}

	s.linkRepo.IncrementClickCount(link.ID)

	ctx := context.Background()
	counterKey := fmt.Sprintf("link:%s:clicks", shortCode)
	s.redisClient.Incr(ctx, counterKey)

	return nil
}

func (s *LinkService) GetDashboardStats(userID int64) (*models.DashboardStats, error) {
	stats, err := s.linkRepo.GetDashboardStats(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve statistics")
	}

	return stats, nil
}
