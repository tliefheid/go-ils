package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/yourusername/library-ils-backend/internal/model"
)

// --- Member Handlers ---
func (s *Service) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	members, err := s.repository.ListMemberss()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, members)
}
func (s *Service) getMemberHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing member ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	member, err := s.repository.GetMember(id)
	if err != nil {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	writeJSON(w, member)
}

func (s *Service) addMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.repository.AddMember(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add member: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, m)
}

func (s *Service) editMemberHandler(w http.ResponseWriter, r *http.Request) {
	var m model.Member

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == 0 {
		http.Error(w, "Missing member ID", http.StatusBadRequest)
		return
	}

	err = s.repository.UpdateMember(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	err = s.repository.DeleteMember(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete member: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
