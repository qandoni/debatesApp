package posts_http_transport

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type PostDTOResponse struct {
	ID        int
	Version   int
	AuthorID  int
	Content   string
	IsDebate  bool
	Images    []domain.PostImage
	Debate    *domain.Debate
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

func postDTOFromDomain(post domain.Post) PostDTOResponse {
	return PostDTOResponse{
		ID:        post.ID,
		Version:   post.Version,
		AuthorID:  post.AuthorID,
		Content:   post.Content,
		IsDebate:  post.IsDebate,
		Images:    post.Images,
		Debate:    post.Debate,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		DeletedAt: post.DeletedAt,
	}
}

func postsDTOFromDomains(posts []domain.Post) []PostDTOResponse {
	postsDTO := make([]PostDTOResponse, len(posts))
	for i, post := range posts {
		postsDTO[i] = postDTOFromDomain(post)
	}
	return postsDTO
}
