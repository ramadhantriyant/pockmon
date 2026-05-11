package util

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ramadhantriyant/pockmon/internal/database"
)

var categories = []struct {
	name string
	typ  string
}{
	{"Salary", "income"},
	{"Interest", "income"},
	{"Dividend", "income"},
	{"Food", "expense"},
	{"Grocery", "expense"},
	{"Transportation", "expense"},
	{"Streaming", "expense"},
	{"Cloud Services", "expense"},
	{"Tax", "expense"},
}

func SeedCategory(ctx context.Context, querier database.Querier, userID pgtype.UUID) error {
	for _, category := range categories {
		categoryID, err := GenerateUUID()
		if err != nil {
			return err
		}

		// Create system category
		_, err = querier.CreateCategory(ctx, database.CreateCategoryParams{
			ID:               categoryID,
			UserID:           userID,
			Name:             category.name,
			Type:             category.typ,
			Color:            nil,
			Icon:             nil,
			ParentCategoryID: pgtype.UUID{},
		})
		if err != nil {
			return err
		}

		// Set system category
		if err := querier.SetSystemCategory(ctx, database.SetSystemCategoryParams{
			ID:     categoryID,
			UserID: userID,
		}); err != nil {
			return err
		}
	}

	return nil
}
