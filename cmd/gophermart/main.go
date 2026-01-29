package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	database "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ptalbayeva/go-musthave-diploma/internal/config"
	"github.com/ptalbayeva/go-musthave-diploma/internal/handler"
	"github.com/ptalbayeva/go-musthave-diploma/internal/postgres"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal("Ошибка конфигурации:", err)
	}

	db, err := connectToDatabase(config.DatabaseURI)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatal("Ошибка выполнения миграций:", err)
	}

	repos := &repository.Repositories{
		Users:   postgres.NewUsersRepository(db),
		Orders:  postgres.NewOrderRepository(db),
		Loyalty: postgres.NewLoyaltyRepository(db),
	}

	authService := service.NewAuthService(repos.Users, config.SecretKey)
	loyaltyService := service.NewLoyaltyService(repos.Loyalty, repos.Orders)

	userHandler := handler.NewUserHandler(authService, loyaltyService, config.SecretKey)

	r := chi.NewRouter()

	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)
	r.Post("/api/user/orders", userHandler.CreateOrder)
	r.Get("/api/user/orders", userHandler.GetOrders)
	r.Get("/api/user/balance", userHandler.GetBalance)
	r.Post("/api/user/balance/withdraw", userHandler.WithdrawPoints)
	r.Get("/api/user/balance/withdrawals", userHandler.GetWithdrawals)

	fmt.Println("Сервер запущен на", config.RunAddress)
	if fail := http.ListenAndServe(config.RunAddress, r); fail != nil {
		log.Fatal("Ошибка запуска сервера:", fail)
	}
}

func connectToDatabase(databaseURI string) (*sql.DB, error) {
	if databaseURI == "" {
		return nil, fmt.Errorf("DATABASE_URI is not set")
	}

	// Открываем подключение к базе данных
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := database.WithInstance(db, &database.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	return nil
}
