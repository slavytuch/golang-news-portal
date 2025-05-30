package main

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

func OpenDatabaseConnection(host string, port string, user string, password string, database string) (*sql.DB, error) {
	dbConfig := mysql.NewConfig()

	dbConfig.Addr = host + ":" + port
	dbConfig.User = user
	dbConfig.Passwd = password
	dbConfig.Net = "tcp"
	dbConfig.DBName = database

	db, err := sql.Open("mysql", dbConfig.FormatDSN())

	if err != nil {
		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		return nil, pingErr
	}

	return db, nil
}
