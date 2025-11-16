package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	TeamID    uuid.UUID
	IsActive  bool
	CreatedAt time.Time
}
