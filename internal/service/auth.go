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

type AuthService struct {
	usersRepo repository.UsersRepository
	SecretKey string
}

func NewAuthService(usersRepo repository.UsersRepository, secretKey string) *AuthService {
	return &AuthService{
		usersRepo: usersRepo,
		SecretKey: secretKey,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	// Хэшируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	// Добавляем пользователя в репозиторий
	userID, err := s.usersRepo.Create(ctx, email, string(passwordHash))
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// AuthenticateUser аутентифицирует пользователя
func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("incorrect password")
	}

	// Генерация JWT токена
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateToken генерирует JWT токен для пользователя
func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.SecretKey))
}
