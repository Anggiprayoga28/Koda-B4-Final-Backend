package handler

import (
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/service"

	"koda-shortlink-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserRegisterRequest true "Registration data"
// @Success 201 {object} response.Response{data=models.AuthResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input", err.Error())
		return
	}

	authResponse, err := h.authService.Register(&req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "Registration successful", authResponse)
}

// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserLoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=models.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input", err.Error())
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "Login successful", authResponse)
}

// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} response.Response{data=models.TokenPair}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh token is required", err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "Token refreshed successfully", tokens)
}

// @Summary Logout user
// @Description Invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh token is required", err.Error())
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		response.BadRequest(c, "Failed to logout", err.Error())
		return
	}

	response.OK(c, "Logout successful", nil)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	response.OK(c, "Google OAuth not implemented yet", nil)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	response.OK(c, "Google OAuth callback not implemented yet", nil)
}
