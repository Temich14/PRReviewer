package dto

import "PRReviewer/internal/core/enums"

type ErrorResponse struct {
	Code    enums.Code `json:"code"`
	Message string     `json:"message"`
}
