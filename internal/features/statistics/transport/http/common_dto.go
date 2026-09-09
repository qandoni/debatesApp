package statistics_http_transport

import "github.com/qandoni/debatesApp/internal/core/domain"

type TopArgumentDTOResponse struct {
	ID            int     `json:"id"`
	AuthorID      int     `json:"author_id"`
	Content       string  `json:"content"`
	AverageRating float64 `json:"average_rating"`
	RatingsCount  int     `json:"ratings_count"`
}

type DebateSideStatisticsDTOResponse struct {
	SideID                int                     `json:"side_id"`
	SideName              string                  `json:"side_name"`
	ParticipantsCount     int                     `json:"participants_count"`
	VotesPercent          float64                 `json:"votes_percent"`
	ArgumentsCount        int                     `json:"arguments_count"`
	AverageArgumentRating float64                 `json:"average_argument_rating"`
	TopArgument           *TopArgumentDTOResponse `json:"top_argument"`
}

type DebateStatisticsDTOResponse struct {
	DebateID     int                               `json:"debate_id"`
	WinnerSideID *int                              `json:"winner_side_id"`
	Sides        []DebateSideStatisticsDTOResponse `json:"sides"`
}

func NewTopArgumentDTOFromDomain(
	argument *domain.TopArgument,
) *TopArgumentDTOResponse {

	if argument == nil {
		return nil
	}

	return &TopArgumentDTOResponse{
		ID:            argument.ID,
		AuthorID:      argument.AuthorID,
		Content:       argument.Content,
		AverageRating: argument.AverageRating,
		RatingsCount:  argument.RatingsCount,
	}
}

func NewDebateSideStatisticsDTOFromDomain(
	side domain.DebateSideStatistics,
) DebateSideStatisticsDTOResponse {
	return DebateSideStatisticsDTOResponse{
		SideID:                side.SideID,
		SideName:              side.SideName,
		ParticipantsCount:     side.ParticipantsCount,
		VotesPercent:          side.VotesPercent,
		ArgumentsCount:        side.ArgumentsCount,
		AverageArgumentRating: side.AverageArgumentRating,
		TopArgument:           NewTopArgumentDTOFromDomain(side.TopArgument),
	}
}

func NewDebateStatisticsDTOFromDomain(
	statistics domain.DebateStatistics,
) DebateStatisticsDTOResponse {

	sides := make(
		[]DebateSideStatisticsDTOResponse,
		0,
		len(statistics.Sides),
	)

	for _, side := range statistics.Sides {
		sides = append(
			sides,
			NewDebateSideStatisticsDTOFromDomain(side),
		)
	}

	return DebateStatisticsDTOResponse{
		DebateID:     statistics.DebateID,
		WinnerSideID: statistics.WinnerSideID,
		Sides:        sides,
	}
}
