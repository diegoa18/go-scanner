package discover

import (
	"context"
	"go-scanner/internal/discover/core"
	"go-scanner/internal/discover/policy"
)

// re-exportar tipos publicos para mantener compatibilidad
type HostResult = core.HostResult
type Discoverer = core.Discoverer
type Policy = policy.Policy

// re-exportar función principal de orquestac	ión
func Run(ctx context.Context, targets []string, pol Policy) ([]HostResult, error) {
	return core.Run(ctx, targets, pol)
}
