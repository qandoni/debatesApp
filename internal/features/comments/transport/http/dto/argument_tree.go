package comments_dto

import "github.com/qandoni/debatesApp/internal/core/domain"

func BuildArgumentTree(
	arguments []domain.CommentWithRating,
	comments []domain.Comment,
) []CommentDTOResponse {

	type commentNode struct {
		dto      CommentDTOResponse
		children []*commentNode
	}

	nodes := make(map[int]*commentNode, len(comments))

	for _, comment := range comments {
		dto := NewCommentDTOFromDomain(comment)

		nodes[comment.ID] = &commentNode{
			dto:      dto,
			children: make([]*commentNode, 0),
		}
	}

	for _, comment := range comments {
		if comment.ParentCommentID == nil {
			continue
		}

		node, exists := nodes[comment.ID]
		if !exists {
			continue
		}

		parent, exists := nodes[*comment.ParentCommentID]
		if !exists {
			continue
		}

		parent.children = append(parent.children, node)
	}

	var buildDTO func(node *commentNode) CommentDTOResponse

	buildDTO = func(node *commentNode) CommentDTOResponse {
		dto := node.dto

		dto.Replies = make([]CommentDTOResponse, 0, len(node.children))

		for _, child := range node.children {
			dto.Replies = append(
				dto.Replies,
				buildDTO(child),
			)
		}

		return dto
	}

	response := make([]CommentDTOResponse, 0, len(arguments))

	for _, argument := range arguments {
		node, exists := nodes[argument.Comment.ID]
		if !exists {
			continue
		}

		dto := buildDTO(node)

		dto.AverageRating = argument.AverageRating
		dto.RatingsCount = argument.RatingsCount
		dto.AuthorLiked = argument.Comment.AuthorLiked

		response = append(response, dto)
	}

	return response
}
