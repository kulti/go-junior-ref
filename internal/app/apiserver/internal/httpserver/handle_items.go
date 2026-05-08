package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type createItemResp struct {
	Item httpItem `json:"item"`
}

type httpItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

func (s *Server) handeCreateItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("list_id")

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed decode create item request", slog.String("err", err.Error()))
		http.Error(w, "invalid body json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "missed required field: name", http.StatusBadRequest)
		return
	}

	itemID, err := s.app.CreateItem(r.Context(), listID, req.Name)
	if err != nil {
		slog.Error("failed create item", slog.String("err", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := createItemResp{
		Item: httpItem{
			ID:   itemID,
			Name: req.Name,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed send response", slog.String("err", err.Error()))
	}
}

func (s *Server) handeDoneItem(w http.ResponseWriter, r *http.Request) {
	itemID := r.PathValue("item_id")

	item, err := s.app.DoneItem(r.Context(), itemID)
	if err != nil {
		slog.Error("failed done item", slog.String("err", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := createItemResp{
		Item: httpItem{
			ID:   itemID,
			Name: item.Name,
			Done: item.Done,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed send response", slog.String("err", err.Error()))
	}
}
