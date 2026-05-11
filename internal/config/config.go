package config

import (
	firebase "firebase.google.com/go/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ramadhantriyant/pockmon/internal/database"
)

type Config struct {
	DB          *pgxpool.Pool
	Querier     database.Querier
	FirebaseApp *firebase.App
}

func New(pool *pgxpool.Pool, firebaseApp *firebase.App) *Config {
	return &Config{
		DB:          pool,
		Querier:     database.New(pool),
		FirebaseApp: firebaseApp,
	}
}
