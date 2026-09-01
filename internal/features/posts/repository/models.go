package posts_repository

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type PostModel struct {
	ID        int
	Version   int
	AuthorID  int
	Content   string
	IsDebate  bool
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

func postDomainsFromModels(posts []PostModel) []domain.Post {
	postDomains := make([]domain.Post, len(posts))
	for i, post := range posts {
		postDomains[i] = domain.NewPost(
			post.ID,
			post.Version,
			post.AuthorID,
			post.Content,
			post.IsDebate,
			post.CreatedAt,
			post.UpdatedAt,
			post.DeletedAt,
		)
	}
	return postDomains
}
