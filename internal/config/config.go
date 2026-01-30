package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	SecretKey            string
}

func NewConfig() (*Config, error) {
	var (
		runAddress           = flag.String("a", ":8080", "Адрес и порт для запуска сервиса")
		databaseURI          = flag.String("d", "postgres://my_superuser:strongpassword@localhost:5432/gophermart?sslmode=disable", "Адрес подключения к базе данных")
		accrualSystemAddress = flag.String("r", "http://localhost:8081", "Адрес системы расчёта начислений")
		secretKey            = flag.String("s", "secretkey", "Secret key")
	)

	// Разбираем флаги командной строки
	flag.Parse()

	// Читаем из переменных окружения, если флаг не передан
	config := &Config{
		RunAddress:           getEnv("RUN_ADDRESS", *runAddress),
		DatabaseURI:          getEnv("DATABASE_URI", *databaseURI),
		AccrualSystemAddress: getEnv("ACCRUAL_SYSTEM_ADDRESS", *accrualSystemAddress),
		SecretKey:            getEnv("SECRET_KEY", *secretKey),
	}

	// Проверяем, что все необходимые параметры заданы
	if config.RunAddress == "" || config.DatabaseURI == "" {
		return nil, fmt.Errorf("необходимо задать все параметры конфигурации: RUN_ADDRESS, DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS")
	}

	return config, nil
}

// getEnv пытается получить значение из переменной окружения, если оно есть. Если переменная окружения не установлена, возвращает значение по умолчанию.
func getEnv(envVar, defaultValue string) string {
	if val, exists := os.LookupEnv(envVar); exists {
		return val
	}
	return defaultValue
}
