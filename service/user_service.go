package service

import (
	"fmt"

	"github.com/Double-O/Limitd-Backend/domain/request_entity"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/rs/zerolog"

	"github.com/Double-O/Limitd-Backend/domain/entity"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req *request_entity.LoginRequest) (*entity.User, error)
	FindUserByEmail(email string) (*entity.User, error)
}

type userServiceImpl struct {
	db *gorm.DB
}

func NewUserService(mainDB *gorm.DB) UserService {
	return &userServiceImpl{
		db: mainDB,
	}
}

func (userService *userServiceImpl) CreateUser(req *request_entity.LoginRequest) (*entity.User, error) {

	user := entity.NewUser(req)

	result := userService.db.Create(&user)
	if result.Error != nil {
		logger.LogMessage(zerolog.ErrorLevel, "service.user_service", "CreateUser", fmt.Sprintf("Error while creating user in db, err : %+v", result.Error))
		return nil, result.Error
	}
	return user, nil
}

func (userService *userServiceImpl) FindUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	result := userService.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		logger.LogMessage(zerolog.ErrorLevel, "service.user_service", "CreateUser", fmt.Sprintf("Error while querying userByEmail in db, err : %+v", result.Error))
		return nil, result.Error
	}
	return &user, nil
}
