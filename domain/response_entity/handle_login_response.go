package response_entity

import (
	"time"

	"github.com/Double-O/Limitd-Backend/domain/entity"

	"github.com/google/uuid"
)

type HandleLongUserResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email" binding:"required"`
	From      string    `json:"from"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HandleLoginResponse struct {
	User        *HandleLongUserResponse `json:"user"`
	AccessToken string                  `json:"access_token"`
}

func ConvertUserToUserResponse(user *entity.User) *HandleLongUserResponse {
	return &HandleLongUserResponse{
		UUID:      user.UUID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		From:      user.From,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
