package util

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GenerateUUID() (pgtype.UUID, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return pgtype.UUID{}, err
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(newUUID.String())

	return pgUUID, err
}

func GetUUID(id string) (pgtype.UUID, error) {
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(id)

	return pgUUID, err
}
