package auth

import (
	"context"
	"project/internal/entity"
)

type User interface {
	GetByLogin(ctx context.Context, login string) (entity.User, error)
}
