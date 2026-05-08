package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"sync/atomic"

	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type app interface {
	CreateList(ctx context.Context, name string) (listID string, err error)
	CreateItem(ctx context.Context, listID, name string) (itemID string, err error)
	GetList(ctx context.Context, listID string) (list models.List, err error)
	DoneItem(ctx context.Context, itemID string) (item models.Item, err error)
	Subscribe(ctx context.Context, listID, email string) error
}

type Server struct {
	mux   *http.ServeMux
	app   app
	addr  string
	ready atomic.Bool
}

type Params struct {
	ListenAddress string
	App           app
	ReadyHandler  http.HandlerFunc
}

func New(p Params) *Server {
	s := &Server{mux: http.NewServeMux(), app: p.App, addr: p.ListenAddress}
	s.mux.Handle("GET /_/ready", p.ReadyHandler)
	s.mux.HandleFunc("GET /debug/info", handleServiceInfo)
	s.mux.HandleFunc("POST /v1/lists", s.handeCreateList)
	s.mux.HandleFunc("GET /v1/lists/{list_id}", s.handeGetList)
	s.mux.HandleFunc("POST /v1/lists/{list_id}/items", s.handeCreateItem)
	s.mux.HandleFunc("POST /v1/lists/{list_id}/items/{item_id}/done", s.handeDoneItem)
	s.mux.HandleFunc("POST /v1/lists/{list_id}/subscribe", s.handleSubscribe)
	return s
}

func (s *Server) Run(ctx context.Context) {
	var ln net.Listener
	for {
		var err error
		lc := net.ListenConfig{}
		ln, err = lc.Listen(ctx, "tcp", s.addr)
		if err != nil {
			slog.Warn("failed to listen", slog.String("addr", s.addr), slog.String("err", err.Error()))
			continue
		}
		break
	}

	server := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	s.ready.Store(true)

	go func() {
		<-ctx.Done()
		s.ready.Store(false)
		if err := server.Close(); err != nil {
			slog.Warn("failed to close server", slog.String("err", err.Error()))
		}
	}()

	if err := server.Serve(ln); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		slog.Warn("failed to serve", slog.String("err", err.Error()))
	}
}

func (s *Server) Ready() error {
	if !s.ready.Load() {
		return errNotReadyToAcceptConnections
	}
	return nil
}

func handleServiceInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("full") {
		fmt.Fprintln(w, buildInfo.String())
	} else {
		var vcsRevision, vcsTime string
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				vcsRevision = setting.Value
			case "vcs.time":
				vcsTime = setting.Value
			}
		}
		fmt.Fprintln(w, "Go Version:", buildInfo.GoVersion, "Build Version:", vcsRevision[:8],
			"Build Time:", vcsTime)
	}
}

//nolint:gochecknoglobals // read-only variable
var buildInfo *debug.BuildInfo = func() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return info
}()
