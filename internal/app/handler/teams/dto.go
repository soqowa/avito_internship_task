package teams

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}
