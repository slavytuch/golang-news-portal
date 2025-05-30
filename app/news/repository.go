package news

type NewsRepository interface {
	All() ([]NewsItem, error)
	Load(id int) (*NewsItem, error)
	Create(title string, body string, image *string) (*int, error)
	Save(item NewsItem) error
	Delete(id int) error
}
