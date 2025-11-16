package users

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User User `json:"user"`
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
	UserIDs []string `json:"userIds"`
}

type BulkDeactivateUsersResponse struct {
	DeactivatedCount     int `json:"deactivatedCount"`
	ReassignedSlotsCount int `json:"reassignedSlotsCount"`
}
