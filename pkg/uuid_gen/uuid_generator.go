package uuidgen

import (
	"github.com/google/uuid"
)

func UUIDGen() string {
	return uuid.Must(uuid.NewV7()).String()
}
