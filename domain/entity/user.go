package entity

import (
	"github.com/Double-O/Limitd-Backend/domain/request_entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email" binding:"required"`
	From      string    `json:"from"`
}

func (User) TableName() string {
	return "user"
}

func NewUser(req *request_entity.LoginRequest) *User {
	user := User{
		UUID:      uuid.New(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		From:      req.From,
	}
	return &user
}
