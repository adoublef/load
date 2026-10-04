package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		fmt.Printf("run: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var c serveCmd
	return c.run(ctx)
}
