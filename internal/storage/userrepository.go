package storage

import (
	"database/sql"
	"github.com/loveletter4you/user-segmentation-service/internal/model"
)

type UserRepository struct {
	storage *Storage
}

func (sr *UserRepository) CreateUser(tx *sql.Tx, user *model.User) error {
	query := "INSERT INTO users DEFAULT VALUES RETURNING id"
	err := sr.storage.DoQueryRow(tx, query).Scan(&user.Id)
	return err
}

func (sr *UserRepository) GetUsers(tx *sql.Tx) ([]*model.User, error) {
	users := make([]*model.User, 0)
	query := "SELECT id FROM users"
	rows, err := sr.storage.DoQuery(tx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(&user.Id); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, err
}
