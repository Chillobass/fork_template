//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package create

import (
	"context"

	"github.com/Chillobass/fork_template/internal/domain/trip"
)

type usecase interface {
	Create(ctx context.Context, input trip.Data) (trip.Trip, error)
}
