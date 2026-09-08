package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"ticket-system/internal/auth"
	"ticket-system/internal/models"
	"ticket-system/internal/repository"
)

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrUserAlreadyExists  = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}

	if len(req.Password) < 6 {
		return nil, ErrPasswordTooShort
	}

	// Check existing
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return &models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, 24)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
	}, nil
}
