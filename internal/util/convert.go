package util

import (
	"strconv"

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

func GetNumeric(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(f, 'f', -1, 64))

	return n, err
}

func GetDate(s string) (pgtype.Date, error) {
	var d pgtype.Date
	err := d.Scan(s)

	return d, err
}
