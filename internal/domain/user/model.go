package user

import (
	"time"
)

type User struct {
	ID        string
	Name      string
	TeamID    string
	IsActive  bool
	CreatedAt time.Time
}
