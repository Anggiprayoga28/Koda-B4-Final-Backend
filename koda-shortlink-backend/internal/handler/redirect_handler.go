package handler

import (
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/service"
	"koda-shortlink-backend/internal/utils"
	"koda-shortlink-backend/pkg/response"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RedirectHandler struct {
	linkService *service.LinkService
	redisClient *redis.Client
}

func NewRedirectHandler(linkService *service.LinkService, redisClient *redis.Client) *RedirectHandler {
	return &RedirectHandler{
		linkService: linkService,
		redisClient: redisClient,
	}
}

// @Summary Redirect to destination
// @Description Redirect to the original URL and log analytics
// @Tags redirect
// @Param shortCode path string true "Short code"
// @Success 302 "Redirect to destination URL"
// @Failure 404 {object} response.Response
// @Router /{shortCode} [get]
func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")

	destination, err := h.linkService.GetDestination(shortCode)
	if err != nil {
		response.NotFound(c, "Link not found or expired")
		return
	}

	deviceInfo := utils.ParseUserAgent(c.Request.UserAgent())

	referer := c.Request.Referer()
	var refererPtr *string
	if referer != "" {
		refererPtr = &referer
	}

	click := &models.Click{
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Referer:    refererPtr,
		DeviceType: &deviceInfo.DeviceType,
		Browser:    &deviceInfo.Browser,
		OS:         &deviceInfo.OS,
		Country:    nil,
		City:       nil,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in analytics logging: %v", r)
			}
		}()

		if err := h.linkService.RecordClick(shortCode, click); err != nil {
			log.Printf("Failed to record click for %s: %v", shortCode, err)
		}
	}()

	c.Redirect(302, destination)
}
