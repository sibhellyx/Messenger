package userrepo

import (
	"strings"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUsers(search string) ([]*entity.User, error) {
	var users []*entity.User

	query := r.db.Model(&entity.User{})

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(surname) LIKE ? OR LOWER(tgname) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	result := query.Select("id, name, surname, tgname, created_at, updated_at").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
