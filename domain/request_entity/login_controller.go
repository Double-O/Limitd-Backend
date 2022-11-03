package request_entity

type LoginRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" binding:"required"`
	From      string `json:"from"`
	Token     string `json:"token" binding:"required"`
}
