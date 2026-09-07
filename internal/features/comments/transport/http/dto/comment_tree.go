package comments_dto

func BuildCommentTree(
	comments []CommentDTOResponse,
) []CommentDTOResponse {
	commentsByID := make(map[int]*CommentDTOResponse, len(comments))
	for _, comment := range comments {
		commentsByID[comment.ID] = &comment
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
