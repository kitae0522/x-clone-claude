package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
}

type authService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	expiryHours int
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, expiryHours int) AuthService {
	return &authService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		expiryHours: expiryHours,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal("failed to check email")
	}
	if exists {
		return nil, apperror.Conflict("email already exists")
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperror.Internal("failed to check username")
	}
	if exists {
		return nil, apperror.Conflict("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, apperror.Internal("failed to hash password")
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Username:     req.Username,
		DisplayName:  req.Username,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Internal("failed to create user")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to generate token")
	}

	userResp := dto.ToUserResponse(user)
	return &dto.AuthResponse{User: userResp, Token: token}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.Unauthorized("invalid email or password")
		}
		return nil, apperror.Internal("failed to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to generate token")
	}

	userResp := dto.ToUserResponse(user)
	return &dto.AuthResponse{User: userResp, Token: token}, nil
}

func (s *authService) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.Unauthorized("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	userResp := dto.ToUserResponse(user)
	return &userResp, nil
}

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(s.expiryHours) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
