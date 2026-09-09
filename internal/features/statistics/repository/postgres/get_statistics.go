package statistics_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *DebateStatisticsRepository) GetStatistics(
	ctx context.Context,
	debateID int,
) (domain.DebateStatistics, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
		SELECT
			ds.id AS side_id,
			ds.name AS side_name,
			d.post_id,
			d.winner_side_id,

			(
				SELECT COUNT(*)
				FROM debatesapp.debate_votes dv
				WHERE dv.debate_id = d.id
				  AND dv.debate_side_id = ds.id
			) AS participants_count,

			COALESCE(
				(
					SELECT COUNT(*)
					FROM debatesapp.debate_votes dv
					WHERE dv.debate_id = d.id
					  AND dv.debate_side_id = ds.id
				) * 100.0
				/
				NULLIF(
					(
						SELECT COUNT(*)
						FROM debatesapp.debate_votes dv_total
						WHERE dv_total.debate_id = d.id
					),
					0
				),
				0
			) AS votes_percent,

			(
				SELECT COUNT(*)
				FROM debatesapp.comments c
				WHERE c.post_id = d.post_id
				  AND c.debate_side_id = ds.id
				  AND c.parent_comment_id IS NULL
			) AS arguments_count,

			COALESCE(
				(
					SELECT AVG(argument_rating.average_rating)
					FROM (
						SELECT
							c.id,
							COALESCE(AVG(cr.score), 0) AS average_rating
						FROM debatesapp.comments c
						LEFT JOIN debatesapp.comment_ratings cr
							ON cr.comment_id = c.id
						WHERE c.post_id = d.post_id
						  AND c.debate_side_id = ds.id
						  AND c.parent_comment_id IS NULL
						GROUP BY c.id
					) AS argument_rating
				),
				0
			) AS average_argument_rating

		FROM debatesapp.debate_sides ds
		JOIN debatesapp.debates d
			ON d.id = ds.debate_id

		WHERE ds.debate_id = $1

		ORDER BY ds.display_order, ds.id
	`

	rows, err := db.Query(ctx, query, debateID)
	if err != nil {
		return domain.DebateStatistics{}, fmt.Errorf(
			"get debate statistics: %w",
			err,
		)
	}
	defer rows.Close()

	var (
		statistics  domain.DebateStatistics
		initialized bool
	)

	statistics.DebateID = debateID
	statistics.Sides = make([]domain.DebateSideStatistics, 0)

	for rows.Next() {
		var (
			sideID                int
			sideName              string
			postID                int
			winnerSideID          *int
			participantsCount     int
			votesPercent          float64
			argumentsCount        int
			averageArgumentRating float64
		)

		err := rows.Scan(
			&sideID,
			&sideName,
			&postID,
			&winnerSideID,
			&participantsCount,
			&votesPercent,
			&argumentsCount,
			&averageArgumentRating,
		)
		if err != nil {
			return domain.DebateStatistics{}, fmt.Errorf(
				"scan debate statistics: %w",
				err,
			)
		}

		if !initialized {
			statistics.WinnerSideID = winnerSideID
			initialized = true
		}

		var topArgument *domain.TopArgument

		if argumentsCount > 0 {
			topArgument, err = r.getTopArgument(
				ctx,
				postID,
				sideID,
			)
			if err != nil {
				return domain.DebateStatistics{}, fmt.Errorf(
					"get top argument for side %d: %w",
					sideID,
					err,
				)
			}
		}

		sideStatistics := domain.DebateSideStatistics{
			SideID:                sideID,
			SideName:              sideName,
			ParticipantsCount:     participantsCount,
			VotesPercent:          votesPercent,
			ArgumentsCount:        argumentsCount,
			AverageArgumentRating: averageArgumentRating,
			TopArgument:           topArgument,
		}

		statistics.Sides = append(
			statistics.Sides,
			sideStatistics,
		)
	}

	if err := rows.Err(); err != nil {
		return domain.DebateStatistics{}, fmt.Errorf(
			"iterate debate statistics: %w",
			err,
		)
	}

	if !initialized {
		return domain.DebateStatistics{}, core_errors.ErrNotFound
	}

	return statistics, nil
}
