package handler

import (
	"koda-shortlink-backend/internal/middleware"
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/service"
	"koda-shortlink-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkService *service.LinkService
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{linkService: linkService}
}

// @Summary Create short link
// @Description Create a new short link (authentication optional)
// @Tags links
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateLinkRequest true "Link data"
// @Success 201 {object} response.Response{data=models.LinkResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req models.CreateLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input", err.Error())
		return
	}

	var userID *int64
	if id, exists := middleware.GetUserID(c); exists {
		userID = &id
	}

	link, err := h.linkService.CreateLink(&req, userID)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "Link created successfully", link)
}

// @Summary Get user links
// @Description Get all links created by the authenticated user
// @Tags links
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=models.LinkListResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/links [get]
func (h *LinkHandler) GetUserLinks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	links, err := h.linkService.GetUserLinks(userID, page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	response.OK(c, "Links retrieved successfully", links)
}

// @Summary Get link details
// @Description Get details of a specific link by short code
// @Tags links
// @Produce json
// @Security BearerAuth
// @Param shortCode path string true "Short code"
// @Success 200 {object} response.Response{data=models.LinkResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/links/{shortCode} [get]
func (h *LinkHandler) GetLinkByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	link, err := h.linkService.GetLinkByShortCode(shortCode, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "Link retrieved successfully", link)
}

// @Summary Update link
// @Description Update link details
// @Tags links
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param shortCode path string true "Short code"
// @Param request body models.UpdateLinkRequest true "Update data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/links/{shortCode} [put]
func (h *LinkHandler) UpdateLink(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var req models.UpdateLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input", err.Error())
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if err := h.linkService.UpdateLink(shortCode, userID, &req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.OK(c, "Link updated successfully", nil)
}

// @Summary Delete link
// @Description Delete a link by short code
// @Tags links
// @Produce json
// @Security BearerAuth
// @Param shortCode path string true "Short code"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/links/{shortCode} [delete]
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	shortCode := c.Param("shortCode")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if err := h.linkService.DeleteLink(shortCode, userID); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.OK(c, "Link deleted successfully", nil)
}

// @Summary Get dashboard stats
// @Description Get statistics for the authenticated user
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.DashboardStats}
// @Failure 401 {object} response.Response
// @Router /api/v1/dashboard/stats [get]
func (h *LinkHandler) GetDashboardStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	stats, err := h.linkService.GetDashboardStats(userID)
	if err != nil {
		response.InternalServerError(c, err.Error(), nil)
		return
	}

	response.OK(c, "Statistics retrieved successfully", stats)
}
