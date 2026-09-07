package comments_dto

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type CommentDTOResponse struct {
	ID              int                  `json:"id"`
	PostID          int                  `json:"post_id"`
	ParentCommentID *int                 `json:"parent_comment_id"`
	AuthorID        int                  `json:"author_id"`
	DebateSideID    *int                 `json:"debate_side_id"`
	Content         string               `json:"content"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       *time.Time           `json:"updated_at"`
	Replies         []CommentDTOResponse `json:"replies"`
}

func NewCommentDTOFromDomain(comment domain.Comment) CommentDTOResponse {
	return CommentDTOResponse{
		ID:              comment.ID,
		PostID:          comment.PostID,
		ParentCommentID: comment.ParentCommentID,
		AuthorID:        comment.AuthorID,
		DebateSideID:    comment.DebateSideID,
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
		Replies:         make([]CommentDTOResponse, 0),
	}
}
