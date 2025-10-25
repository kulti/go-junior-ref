package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type createListResp struct {
	List httpList `json:"list"`
}

type getListResp struct {
	List httpList `json:"list"`
}

type httpList struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []httpItem `json:"items,omitempty"`
}

func (s *Server) handeCreateList(w http.ResponseWriter, r *http.Request) {
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

	listID, err := s.app.CreateList(r.Context(), req.Name)
	if err != nil {
		slog.Error("failed create list", slog.String("err", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := createListResp{
		List: httpList{
			ID:   listID,
			Name: req.Name,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// FIXME: how to handle this error?
	}
}

func (s *Server) handeGetList(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("list_id")
	list, err := s.app.GetList(r.Context(), listID)
	if err != nil {
		slog.Error("failed get list", slog.String("err", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	items := make([]httpItem, len(list.Items))
	for i, v := range list.Items {
		items[i] = httpItem{
			ID:   v.ID,
			Name: v.Name,
			Done: v.Done,
		}
	}

	resp := getListResp{
		List: httpList{
			ID:    list.ID,
			Name:  list.Name,
			Items: items,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// FIXME: how to handle this error?
	}
}
