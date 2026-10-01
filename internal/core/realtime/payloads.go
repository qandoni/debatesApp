package core_realtime

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type CommentCreatedData struct {
	CommentID       int       `json:"comment_id"`
	PostID          int       `json:"post_id"`
	ParentCommentID *int      `json:"parent_comment_id"`
	AuthorID        int       `json:"author_id"`
	DebateSideID    *int      `json:"debate_side_id"`
	AuthorLiked     bool      `json:"author_liked"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
}

func NewCommentCreatedData(comment domain.Comment) CommentCreatedData {
	return CommentCreatedData{
		CommentID:       comment.ID,
		PostID:          comment.PostID,
		ParentCommentID: comment.ParentCommentID,
		AuthorID:        comment.AuthorID,
		DebateSideID:    comment.DebateSideID,
		AuthorLiked:     comment.AuthorLiked,
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt,
	}
}

type CommentUpdatedData struct {
	CommentID       int        `json:"comment_id"`
	PostID          int        `json:"post_id"`
	ParentCommentID *int       `json:"parent_comment_id"`
	DebateSideID    *int       `json:"debate_side_id"`
	AuthorID        int        `json:"author_id"`
	Content         string     `json:"content"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

func NewCommentUpdatedData(comment domain.Comment) CommentUpdatedData {
	return CommentUpdatedData{
		CommentID:       comment.ID,
		PostID:          comment.PostID,
		ParentCommentID: comment.ParentCommentID,
		DebateSideID:    comment.DebateSideID,
		AuthorID:        comment.AuthorID,
		Content:         comment.Content,
		UpdatedAt:       comment.UpdatedAt,
	}
}

type AuthorLikeData struct {
	CommentID      int  `json:"comment_id"`
	PostID         int  `json:"post_id"`
	DebateAuthorID int  `json:"debate_author_id"`
	AuthorLiked    bool `json:"author_liked"`
}

func NewAuthorLikeData(comment domain.Comment, debateAuthorID int) AuthorLikeData {
	return AuthorLikeData{
		CommentID:      comment.ID,
		PostID:         comment.PostID,
		DebateAuthorID: debateAuthorID,
		AuthorLiked:    comment.AuthorLiked,
	}
}

type RatingData struct {
	CommentID int  `json:"comment_id"`
	PostID    int  `json:"post_id"`
	UserID    int  `json:"user_id"`
	Score     int  `json:"score"`
	IsNew     bool `json:"is_new"`
}

func NewRatingData(rating domain.CommentRating, postID int, isNew bool) RatingData {
	return RatingData{
		CommentID: rating.CommentID,
		PostID:    postID,
		UserID:    rating.UserID,
		Score:     rating.Rating,
		IsNew:     isNew,
	}
}

type VoteData struct {
	DebateID     int  `json:"debate_id"`
	PostID       int  `json:"post_id"`
	UserID       int  `json:"user_id"`
	DebateSideID int  `json:"debate_side_id"`
	IsChanged    bool `json:"is_changed"`
}

func NewVoteData(vote domain.DebateVote, postID int, isChanged bool) VoteData {
	return VoteData{
		DebateID:     vote.DebateID,
		PostID:       postID,
		UserID:       vote.UserID,
		DebateSideID: vote.DebateSideID,
		IsChanged:    isChanged,
	}
}

type DebateFinishedData struct {
	DebateID         int  `json:"debate_id"`
	PostID           int  `json:"post_id"`
	WinnerSideID     *int `json:"winner_side_id"`
	FinishedByUserID int  `json:"finished_by_user_id"`
}

func NewDebateFinishedData(debate domain.Debate, finishedByUserID int) DebateFinishedData {
	return DebateFinishedData{
		DebateID:         debate.ID,
		PostID:           debate.PostID,
		WinnerSideID:     debate.WinnerSideID,
		FinishedByUserID: finishedByUserID,
	}
}
