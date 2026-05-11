package config

import (
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ramadhantriyant/pockmon/internal/database"
)

type Config struct {
	DB            *pgxpool.Pool
	Querier       database.Querier
	FirebaseApp   *firebase.App
	AuthClient    *auth.Client
	StorageClient *storage.Client
	StorageBucket string
}

func New(pool *pgxpool.Pool, firebaseApp *firebase.App) *Config {
	return &Config{
		DB:          pool,
		Querier:     database.New(pool),
		FirebaseApp: firebaseApp,
	}
}
