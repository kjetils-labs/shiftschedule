package main

import (
	"context"

	"github.com/shiftschedule/internal/api"
)

func main() {

	ctx := context.Background()
	api.InitHttpServer(ctx)
}
