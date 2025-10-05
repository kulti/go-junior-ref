package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"

	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type app interface {
	CreateList(ctx context.Context, name string) (listID string, err error)
	GetList(ctx context.Context, listID string) (list models.List, err error)
}

type Server struct {
	mux  *http.ServeMux
	app  app
	addr string
}

type Params struct {
	ListenAddress string
	App           app
}

func New(p Params) *Server {
	s := &Server{mux: http.NewServeMux(), app: p.App, addr: p.ListenAddress}
	s.mux.HandleFunc("GET /debug/info", handleServiceInfo)
	s.mux.HandleFunc("POST /v1/lists", s.handeCreateList)
	s.mux.HandleFunc("GET /v1/lists/{list_id}", s.handeGetList)
	return s
}

func (s *Server) Run(ctx context.Context) {
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		if err := server.Close(); err != nil {
			slog.Warn("failed to close server", slog.String("err", err.Error()))
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		slog.Warn("failed to serve", slog.String("err", err.Error()))
	}
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
		fmt.Fprintln(w, "Go Version:", buildInfo.GoVersion, "Build Version:", vcsRevision[:8], "Build Time:", vcsTime)
	}
}

var buildInfo *debug.BuildInfo = func() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()
	return info
}()
