package service

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string) (uuid.UUID, error)
	AuthenticateUser(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	usersRepo repository.UsersRepository
	SecretKey string
}

func NewAuthService(usersRepo repository.UsersRepository, secretKey string) *authService {
	return &authService{
		usersRepo: usersRepo,
		SecretKey: secretKey,
	}
}

func (s *authService) RegisterUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := s.usersRepo.Create(ctx, email, string(passwordHash))
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return uuid.Nil, repository.ErrConflict
		}
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *authService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("incorrect password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.SecretKey))
}
