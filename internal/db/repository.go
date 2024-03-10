package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"mount-service/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(config *model.Config) *UserRepository {
	connInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
	)

	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}

	return &UserRepository{db: db}
}

func (repo *UserRepository) GetUser(username string) *model.User {

	rows, err := repo.db.Query("SELECT * FROM users WHERE users.username = ?", username)
	if err != nil {
		return nil
	}

	defer rows.Close()

	if rows.Next() {
		user := model.User{}
		rows.Scan(user.Username)
	} else {

	}

	panic("Not implemented yet")
}
