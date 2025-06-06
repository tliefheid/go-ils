package frontend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) returnPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing return id", 400)
		return
	}

	resp, err := http.Post(s.uri+"/returns/"+id, "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to fetch return", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		http.Error(w, "Failed to return book", resp.StatusCode)
		return
	}

	s.borrowPage(w, r) // Redirect to borrow page after successful return
}
