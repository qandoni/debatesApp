package comments_dto

import "github.com/qandoni/debatesApp/internal/core/domain"

func BuildCommentTree(
	comments []domain.Comment,
) []CommentDTOResponse {
	commentsByID := make(map[int]*CommentDTOResponse, len(comments))
	for _, comment := range comments {
		dto := NewCommentDTOFromDomain(comment)
		commentsByID[comment.ID] = &dto
	}

	roots := make([]CommentDTOResponse, 0)
	for _, comment := range comments {
		current := commentsByID[comment.ID]

		if comment.ParentCommentID == nil {
			roots = append(roots, *current)
			continue
		}
		parent, exists := commentsByID[*comment.ParentCommentID]
		if !exists {
			continue
		}
		parent.Replies = append(parent.Replies, *current)
	}
	return roots
}
