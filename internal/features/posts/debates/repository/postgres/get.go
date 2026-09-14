package debates_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

const getByIDQuery = `
SELECT id, post_id, status, end_at, created_at, finished_at, winner_side_id
FROM debatesApp.debates
WHERE id = $1
`

const getByPostIDQuery = `
SELECT id, post_id, status, end_at, created_at, finished_at, winner_side_id
FROM debatesApp.debates
WHERE post_id = $1;
`

const getAuthorIDQuery = `
SELECT p.author_id
FROM debatesApp.debates d
JOIN debatesApp.posts p ON p.id = d.post_id
WHERE d.id = $1
`

func (r *DebatesRepository) GetByID(
	ctx context.Context,
	debateID int,
) (domain.Debate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, getByIDQuery, debateID)
	var debateModel DebateModel

	err := row.Scan(
		&debateModel.ID,
		&debateModel.PostID,
		&debateModel.Status,
		&debateModel.EndAt,
		&debateModel.CreatedAt,
		&debateModel.FinishedAt,
		&debateModel.WinnerSideID,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Debate{}, fmt.Errorf("debate with id='%d': %w", debateID, core_errors.ErrNotFound)
		}
		return domain.Debate{}, fmt.Errorf("scan error: %w", err)
	}

	debateDomain := domain.NewDebate(
		debateModel.ID,
		debateModel.PostID,
		debateModel.Status,
		debateModel.EndAt,
		debateModel.CreatedAt,
		debateModel.FinishedAt,
		debateModel.WinnerSideID,
	)
	return debateDomain, nil
}

func (r *DebatesRepository) GetByPostID(
	ctx context.Context,
	postID int,
) (domain.Debate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, getByPostIDQuery, postID)
	var debateModel DebateModel

	err := row.Scan(
		&debateModel.ID,
		&debateModel.PostID,
		&debateModel.Status,
		&debateModel.EndAt,
		&debateModel.CreatedAt,
		&debateModel.FinishedAt,
		&debateModel.WinnerSideID,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Debate{}, fmt.Errorf("post with id='%d': %w", postID, core_errors.ErrNotFound)
		}
		return domain.Debate{}, fmt.Errorf("scan error: %w", err)
	}

	debateDomain := domain.NewDebate(
		debateModel.ID,
		debateModel.PostID,
		debateModel.Status,
		debateModel.EndAt,
		debateModel.CreatedAt,
		debateModel.FinishedAt,
		debateModel.WinnerSideID,
	)
	return debateDomain, nil

}

func (r *DebatesRepository) GetAuthorID(
	ctx context.Context,
	debateID int,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var authorID int

	err := db.QueryRow(ctx, getAuthorIDQuery, debateID).Scan(&authorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, core_errors.ErrNotFound
		}
		return 0, fmt.Errorf("get debate author id: %w", err)
	}
	return authorID, nil
}
