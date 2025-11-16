package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	TeamID    uuid.UUID `json:"teamId"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive,omitempty"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
}

type BulkDeactivateUsersRequest struct {
	UserIDs []uuid.UUID `json:"userIds"`
}

type BulkDeactivateUsersResponse struct {
	DeactivatedCount     int `json:"deactivatedCount"`
	ReassignedSlotsCount int `json:"reassignedSlotsCount"`
}
