package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nilebox/k8s-deploy/pkg/app"

	"k8s.io/client-go/rest"
)

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		log.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)
	return runWithEnvConfig(ctx)
	//return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	//config.UserAgent = "k8s-deploy/" + main.Version + "/" + GitCommit

	return runWithConfig(ctx, config)
}

// For minikube local testing only
func runWithEnvConfig(ctx context.Context) error {
	config := configFromEnv()
	//config.UserAgent = "k8s-deploy/" + Version + "/" + GitCommit

	return runWithConfig(ctx, config)
}

func runWithConfig(ctx context.Context, config *rest.Config) error {
	server := app.Server{
		RestConfig: config,
	}
	return server.Run(ctx)
}

// cancelOnInterrupt calls f when os.Interrupt or SIGTERM is received.
// It ignores subsequent interrupts on purpose - program should exit correctly after the first signal.
func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			f()
		}
	}()
}

func configFromEnv() *rest.Config {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		panic("Unable to load cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}
	return &rest.Config{
		Host: "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:   os.Getenv("KUBERNETES_CA_PATH"),
			CertFile: os.Getenv("KUBERNETES_CLIENT_CERT"),
			KeyFile:  os.Getenv("KUBERNETES_CLIENT_KEY"),
		},
	}
}
