package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUsersRepo struct {
	CreateFn     func(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
	GetByEmailFn func(ctx context.Context, email string) (*models.User, error)
}

func (m *mockUsersRepo) Create(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	return m.CreateFn(ctx, email, passwordHash)
}

func (m *mockUsersRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.GetByEmailFn(ctx, email)
}

var errDB = errors.New("db error")

func TestAuthService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		createFn      func(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
		expectedError error
	}{
		{
			name: "success",
			createFn: func(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
				require.NotEmpty(t, passwordHash)
				return uuid.New(), nil
			},
			expectedError: nil,
		},
		{
			name: "conflict",
			createFn: func(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
				return uuid.Nil, repository.ErrConflict
			},
			expectedError: repository.ErrConflict,
		},
		{
			name: "internal error",
			createFn: func(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
				return uuid.Nil, errDB
			},
			expectedError: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUsersRepo{
				CreateFn: tt.createFn,
			}

			svc := service.NewAuthService(repo, "secret")

			_, err := svc.RegisterUser(context.Background(), "test@mail.com", "password")

			if tt.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	password := "secret"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	userID := uuid.New()

	tests := []struct {
		name          string
		getByEmailFn  func(ctx context.Context, email string) (*models.User, error)
		inputPassword string
		expectedError bool
	}{
		{
			name: "success",
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:           userID,
					PasswordHash: string(hash),
				}, nil
			},
			inputPassword: password,
			expectedError: false,
		},
		{
			name: "user not found",
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			inputPassword: password,
			expectedError: true,
		},
		{
			name: "wrong password",
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:           userID,
					PasswordHash: string(hash),
				}, nil
			},
			inputPassword: "wrong",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUsersRepo{
				GetByEmailFn: tt.getByEmailFn,
			}

			svc := service.NewAuthService(repo, "secret")

			token, err := svc.AuthenticateUser(context.Background(), "test@mail.com", tt.inputPassword)

			if tt.expectedError {
				require.Error(t, err)
				require.Empty(t, token)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
			}
		})
	}
}
