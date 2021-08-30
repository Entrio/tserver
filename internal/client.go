package internal

import (
	"github.com/google/uuid"
	"net"
)

type (
	Client struct {
		ID         uuid.UUID
		Connection net.Conn
	}
)
