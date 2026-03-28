package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func Connect() error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
	if err != nil {
		return err
	}
	DB = conn
	return nil
}

func Close() {
	DB.Close(context.Background())
}
