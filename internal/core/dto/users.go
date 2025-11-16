package dto

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active "`
}

type UserIDQuery struct {
	UserID string `form:"user_id" binding:"required"`
}
