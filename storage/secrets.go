package storage

import (
	"context"
)

type Secrets interface {
	GetToken(ctx context.Context, token string) (map[string]interface{}, error)
	CreateToken(ctx context.Context, data map[string]string) (string, error)
	Close()
	SetToken(token string)
	Revert()
}
