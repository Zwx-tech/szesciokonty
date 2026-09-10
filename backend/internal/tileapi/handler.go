package tileapi

import (
	"encoding/json"
	"net/http"

	"github.com/Zwx-tech/szesciokonty/backend/internal/tile"
)

func Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	def := tile.DefByID(id)
	if def == nil {
		http.Error(w, "tile not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(def)
}
