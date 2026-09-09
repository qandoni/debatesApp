package domain

func NewDebateStatistics(
	debateID int,
	winnerSideID *int,
	sides []DebateSideStatistics,
) *DebateStatistics {
	return &DebateStatistics{
		debateID,
		winnerSideID,
		sides,
	}
}

type DebateStatistics struct {
	DebateID     int
	WinnerSideID *int
	Sides        []DebateSideStatistics
}

func NewDebateSideStatistics(
	sideID int,
	sideName string,
	participantsCount int,
	votesPercent float64,
	argumentsCount int,
	averageArgumentRating float64,
	topArgument *TopArgument,
) *DebateSideStatistics {
	return &DebateSideStatistics{
		sideID,
		sideName,
		participantsCount,
		votesPercent,
		argumentsCount,
		averageArgumentRating,
		topArgument,
	}
}

type DebateSideStatistics struct {
	SideID                int
	SideName              string
	ParticipantsCount     int
	VotesPercent          float64
	ArgumentsCount        int
	AverageArgumentRating float64
	TopArgument           *TopArgument
}

func NewTopArgument(
	iD int,
	authorID int,
	content string,
	averageRating float64,
	ratingsCount int,
) *TopArgument {
	return &TopArgument{
		iD,
		authorID,
		content,
		averageRating,
		ratingsCount,
	}
}

type TopArgument struct {
	ID            int
	AuthorID      int
	Content       string
	AverageRating float64
	RatingsCount  int
}
