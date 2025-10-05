package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("API_SERVER_DB_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	http.HandleFunc("GET /debug/info", handleServiceInfo)
	http.HandleFunc("POST /v1/lists", newHandeCreateList(conn))

	fmt.Println("service run")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Fatal(err)
	}
}

func newHandeCreateList(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed decode create list request", slog.String("err", err.Error()))
			http.Error(w, "invalid body json", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "missed required field: name", http.StatusBadRequest)
			return
		}

		listID := uuid.NewString()
		if _, err := conn.Exec(r.Context(), `INSERT INTO lists(id, name) VALUES($1, $2)`, listID, req.Name); err != nil {
			slog.Error("failed create list in db", slog.String("err", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var resp struct {
			List struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"list"`
		}

		resp.List.ID = listID
		resp.List.Name = req.Name

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// how to handle this error?
		}
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
