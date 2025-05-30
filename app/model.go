package main

type NewsItem struct {
	Id    int
	Title string
	Body  string
	Image *string
}

func LoadNewsItem(id int) (*NewsItem, error) {
	row, err := db.Query("SELECT * FROM news WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	var newsItem NewsItem

	row.Next()
	if err := row.Scan(&newsItem.Id, &newsItem.Title, &newsItem.Body, &newsItem.Image); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &newsItem, nil
}

func CreateNewsItem(title string, body string, image *string) (*int, error) {
	result, err := db.Exec("INSERT INTO news (title, body, image) VALUES (?, ?, ?)", title, body, image)

	if err != nil {
		return nil, err
	}

	insertedId, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	newsId := int(insertedId)

	return &newsId, err
}

func (n *NewsItem) Save() error {
	row, err := db.Query("SELECT * FROM news WHERE id = ?", n.Id)

	if err != nil {
		return err
	}

	var oldNewsItem NewsItem

	row.Next()
	if err := row.Scan(&oldNewsItem.Id, &oldNewsItem.Title, &oldNewsItem.Body, &oldNewsItem.Image); err != nil {
		return err
	}

	_, err = db.Exec("UPDATE news SET title = ?, body = ?, image = ? WHERE id = ?", n.Title, n.Body, n.Image, n.Id)

	if err != nil {
		return err
	}

	if oldNewsItem.Image != nil && n.Image != oldNewsItem.Image {
		DeleteImage(*oldNewsItem.Image)
	}

	return nil
}
