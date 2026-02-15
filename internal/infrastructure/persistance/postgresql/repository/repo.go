package repo

import (
	"context"
	"fmt"
)

type OutBox[T any] struct {
}

func (o *OutBox[T]) Emit(ctx context.Context, event string, data T) error {
	// Logic to save to Postgres or send to NATS
	fmt.Printf("Registering event: %s with data: %+v\n", event, data)

	return nil
}
