package bloc_client

import (
	"github.com/google/uuid"
)

type UUID uuid.UUID

var NillUUID = UUID(uuid.Nil)

func NewUUID() UUID {
	return UUID(uuid.New())
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}
