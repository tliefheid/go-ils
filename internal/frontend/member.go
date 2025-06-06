package frontend

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) memberPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.uri + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}

	defer resp.Body.Close()

	var members []model.Member

	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}

	s.executeTemplate(w, "member.gohtml", members)
}
func (s *Service) memberDetailPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing member id", 400)
		return
	}

	resp, err := http.Get(s.uri + "/members/" + id)
	if err != nil {
		http.Error(w, "Failed to fetch member", 500)
		return
	}

	defer resp.Body.Close()

	var member model.Member

	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		http.Error(w, "Failed to decode member", 500)
		return
	}

	s.executeTemplate(w, "member_detail.gohtml", member)
}
