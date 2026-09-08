package comments_dto

import "github.com/qandoni/debatesApp/internal/core/domain"

func BuildCommentTree(
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

	roots := make([]*commentNode, 0)

	for _, comment := range comments {
		node := nodes[comment.ID]

		if comment.ParentCommentID == nil {
			roots = append(roots, node)
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

	response := make([]CommentDTOResponse, 0, len(roots))

	for _, root := range roots {
		response = append(
			response,
			buildDTO(root),
		)
	}

	return response
}
