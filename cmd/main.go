package main

import (
	"context"
	"fmt"

	"github.com/nilebox/k8s-deploy/pkg/app"

	"k8s.io/client-go/rest"
)

func main() {
	fmt.Println("hello world")
}

func runWithConfig(ctx context.Context, config *rest.Config) error {
	server := app.Server{
		RestConfig: config,
	}
	return server.Run(ctx)
}
