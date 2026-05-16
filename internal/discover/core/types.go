package core

import (
	"context"
	"go-scanner/internal/model"
)

type Discoverer interface {
	Discover(ctx context.Context, target string) (model.HostResult, error)
}
