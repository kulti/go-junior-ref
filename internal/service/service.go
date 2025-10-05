package service

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ServiceComponent describes single service component. Each component works independently.
type ServiceComponent interface {
	// Run implementation should do the following:
	// - uses `ctx` to stop works;
	// - returns from `Run` after stopping completed.
	//
	// `Run` is always call inside new goroutiine.
	// So Run implementation not needed to create an extra goroutine
	// to work independently from other `Run`s.
	Run(ctx context.Context)
}

type Service struct {
	Components []ServiceComponent
}

func Main(newService func() (*Service, error)) {
	s, err := newService()
	if err != nil {
		slog.Error("failed to create service", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(s.Components))
	for _, c := range s.Components {
		go func() {
			defer wg.Done()
			c.Run(ctx)
		}()
	}

	slog.Info("service run")
	wg.Wait()
	slog.Info("service stop")
}
