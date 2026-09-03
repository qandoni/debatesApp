package debates_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *DebatesRepository) GetAuthorID(
	ctx context.Context,
	debateID int,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT p.author_id
	FROM debatesApp.debates d
	JOIN debatesApp.posts p ON p.id = d.post_id
	WHERE d.id = $1
	`

	db := r.dbFromContext(ctx)

	var authorID int

	err := db.QueryRow(ctx, query, debateID).Scan(&authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, core_errors.ErrNotFound
		}
		return 0, fmt.Errorf("get debate author id: %w", err)
	}
	return authorID, nil
}
