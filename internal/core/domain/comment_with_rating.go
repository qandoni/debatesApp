package domain

type CommentWithRating struct {
	Comment       Comment
	AverageRating float64
	RatingsCount  int
}
