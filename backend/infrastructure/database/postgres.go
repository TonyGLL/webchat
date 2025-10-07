package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

func NewDBPool() (*pgxpool.Pool, error) {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development" // default to development
	}

	envFile := fmt.Sprintf("%s.env", env)
    if env == "development" {
        envFile = "dev.env"
    }

	err := godotenv.Load(envFile)
	if err != nil {
		return nil, fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_DATABASE")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return pool, nil
}