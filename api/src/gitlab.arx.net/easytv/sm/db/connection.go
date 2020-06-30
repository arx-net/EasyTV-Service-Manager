package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/lib/pq"
)

type DatabasePool struct {
	DB         *sql.DB
	stmt_cache map[string]*sql.Stmt
}

func Open() (*DatabasePool, error) {
	username := os.Getenv("DB_USER")
	if len(username) == 0 {
		content, err := ioutil.ReadFile(os.Getenv("DB_USER_FILE"))
		if err != nil {
			return nil, err
		}
		username = string(content)
	}

	password := os.Getenv("DB_PASSWORD")
	if len(password) == 0 {
		content, err := ioutil.ReadFile(os.Getenv("DB_PASSWORD_FILE"))
		if err != nil {
			return nil, err
		}
		password = string(content)
	}

	con_str := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable host=%v",
		username,
		password,
		os.Getenv("DB_SCHEMA"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_HOST"))

	db, err := sql.Open("postgres", con_str)

	if err != nil {
		return nil, err
	}
	return &DatabasePool{
		DB:         db,
		stmt_cache: make(map[string]*sql.Stmt),
	}, nil
}

func (this *DatabasePool) Prepare(query string) (*sql.Stmt, error) {
	if stmt, ok := this.stmt_cache[query]; ok {
		return stmt, nil
	}

	stmt, err := this.DB.Prepare(query)

	if err != nil {
		return nil, err
	}

	this.stmt_cache[query] = stmt

	return stmt, nil
}

func (this *DatabasePool) Close() {
	for _, stmt := range this.stmt_cache {
		stmt.Close()
	}
	this.DB.Close()
}
