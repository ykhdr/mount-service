package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"mount-service/internal/model"
)

var log = logrus.New()

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
		log.WithError(err).Panicln("Can't connect to database")
	}

	log.WithFields(logrus.Fields{
		"db_host": config.DBHost,
		"db_user": config.DBUser,
		"db_name": config.DBName,
	}).Infoln("Connect database")

	return &UserRepository{db: db}
}

func (repo *UserRepository) GetUser(username string) *model.User {

	rows, err := repo.db.Query("SELECT * FROM users WHERE users.username=?", username)
	if err != nil {
		log.WithError(err).Panicln("Error on db query")
		return nil
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Warningln("Error on closing query response rows")
		}
	}(rows)

	if !rows.Next() {
		return nil
	}

	user := &model.User{}
	err = rows.Scan(user.Username)
	if err != nil {
		log.WithError(err).Error("Error on scan username from rows")
		return nil
	}

	return user
}
