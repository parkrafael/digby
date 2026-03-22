package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"os"
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
