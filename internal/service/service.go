package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

	// Ready returns nil if components is ready to work.
	// Otherwise returns error with reason.
	Ready() error
}

type Service struct {
	components map[string]ServiceComponent
}

func New() *Service {
	return &Service{components: map[string]ServiceComponent{}}
}

func (s *Service) AddComponent(name string, c ServiceComponent) {
	if _, ok := s.components[name]; ok {
		panic("duplicate component name: " + name)
	}
	s.components[name] = c
}

func (s *Service) HandleReady(w http.ResponseWriter, _ *http.Request) {
	for n, c := range s.components {
		if err := c.Ready(); err != nil {
			http.Error(w, fmt.Sprintf("component %q is not ready: %s", n, err.Error()), http.StatusServiceUnavailable)
			return
		}
	}
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
	wg.Add(len(s.components))
	for n, c := range s.components {
		go func() {
			defer func() {
				slog.Info("component stop", slog.String("name", n))
				wg.Done()
			}()
			slog.Info("component running", slog.String("name", n))
			c.Run(ctx)
			slog.Info("component run", slog.String("name", n))
		}()
	}

	slog.Info("service run")
	wg.Wait()
	slog.Info("service stop")
}
