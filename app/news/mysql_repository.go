package news

import (
	"database/sql"
	"errors"
)

type MysqlNewsRepository struct {
	DB *sql.DB
}

func (r *MysqlNewsRepository) All() ([]NewsItem, error) {
	var newsList []NewsItem

	rows, err := r.DB.Query("SELECT * FROM news")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var newsItem NewsItem

		if err := rows.Scan(&newsItem.Id, &newsItem.Title, &newsItem.Body, &newsItem.Image); err != nil {
			return nil, err
		}

		newsList = append(newsList, newsItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return newsList, nil
}

func (r *MysqlNewsRepository) Create(title string, body string, image *string) (*int, error) {
	result, err := r.DB.Exec("INSERT INTO news (title, body, image) VALUES (?, ?, ?)", title, body, image)

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

func (r *MysqlNewsRepository) Load(id int) (*NewsItem, error) {
	row, err := r.DB.Query("SELECT * FROM news WHERE id = ?", id)

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

func (r *MysqlNewsRepository) Delete(id int) error {
	_, err := r.DB.Exec("DELETE FROM news WHERE id = ?", id)

	if err != nil {
		return err
	}

	return nil
}

func (r *MysqlNewsRepository) Save(n NewsItem) error {
	row, err := r.DB.Query("SELECT count(*) FROM news WHERE id = ?", n.Id)

	if err != nil {
		return err
	}

	if !row.Next() {
		return errors.New("News item not found")
	}

	_, err = r.DB.Exec("UPDATE news SET title = ?, body = ?, image = ? WHERE id = ?", n.Title, n.Body, n.Image, n.Id)

	if err != nil {
		return err
	}

	return nil
}
