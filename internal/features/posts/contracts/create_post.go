package posts_contracts

type CreatePostInput struct {
	AuthorID int
	Content  string
	IsDebate bool
}
