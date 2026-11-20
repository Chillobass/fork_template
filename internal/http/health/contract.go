package health

import "context"

type pinger interface {
	Ping(context.Context) error
}
