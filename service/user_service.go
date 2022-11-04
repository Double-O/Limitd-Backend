package service

import (
	"fmt"

	"github.com/Double-O/Limitd-Backend/domain/custom_error"

	"github.com/Double-O/Limitd-Backend/utils"

	"github.com/Double-O/Limitd-Backend/domain/request_entity"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/rs/zerolog"

	"github.com/Double-O/Limitd-Backend/domain/entity"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req *request_entity.HandleLoginRequest) (*entity.User, *custom_error.Error)
	FindUserByEmail(email string) (*entity.User, *custom_error.Error)
}

type userServiceImpl struct {
	db *gorm.DB
}

func NewUserService(mainDB *gorm.DB) UserService {
	return &userServiceImpl{
		db: mainDB,
	}
}

func (userService *userServiceImpl) CreateUser(req *request_entity.HandleLoginRequest) (*entity.User, *custom_error.Error) {

	user := entity.NewUser(req)

	result := userService.db.Create(&user)
	if result.Error != nil {
		errorMessage := fmt.Sprintf(utils.CreateUserErrorMsg, result.Error)
		customErr := custom_error.NewErrorFromMessage("CreateUserErrorMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "service.user_service", "CreateUser", errorMessage)
		return nil, customErr
	}
	return user, nil
}

func (userService *userServiceImpl) FindUserByEmail(email string) (*entity.User, *custom_error.Error) {
	var user entity.User
	result := userService.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		errorMessage := fmt.Sprintf(utils.FindUserByEmailErrorMsg, result.Error)
		customErr := custom_error.NewErrorFromMessage("FindUserByEmailErrorMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "service.user_service", "CreateUser", errorMessage)
		return nil, customErr
	}
	return &user, nil
}
