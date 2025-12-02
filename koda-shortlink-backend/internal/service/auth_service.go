package service

import (
	"errors"
	"koda-shortlink-backend/internal/models"
	"koda-shortlink-backend/internal/repository"
	"koda-shortlink-backend/internal/utils"
	"time"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtUtil     *utils.JWTUtil
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, jwtUtil *utils.JWTUtil) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtUtil:     jwtUtil,
	}
}

func (s *AuthService) Register(req *models.UserRegisterRequest) (*models.AuthResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	tokens, err := s.generateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.jwtUtil.GetRefreshExpiry()),
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.New("failed to create session")
	}

	return &models.AuthResponse{
		User:   user.ToResponse(),
		Tokens: tokens,
	}, nil
}

func (s *AuthService) Login(req *models.UserLoginRequest) (*models.AuthResponse, error) {
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	valid, err := utils.VerifyPassword(req.Password, user.Password)
	if err != nil || !valid {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	tokens, err := s.generateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.jwtUtil.GetRefreshExpiry()),
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.New("failed to create session")
	}

	return &models.AuthResponse{
		User:   user.ToResponse(),
		Tokens: tokens,
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenPair, error) {
	claims, err := s.jwtUtil.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	session, err := s.sessionRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("session not found or expired")
	}

	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	tokens, err := s.generateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	s.sessionRepo.DeleteByRefreshToken(refreshToken)

	newSession := &models.Session{
		UserID:       session.UserID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.jwtUtil.GetRefreshExpiry()),
	}
	if err := s.sessionRepo.Create(newSession); err != nil {
		return nil, errors.New("failed to create new session")
	}

	return tokens, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.sessionRepo.DeleteByRefreshToken(refreshToken)
}

func (s *AuthService) generateTokenPair(userID int64, email string) (*models.TokenPair, error) {
	accessToken, err := s.jwtUtil.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(userID, email)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
