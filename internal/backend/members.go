package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (s *Service) searchMembers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		fmt.Println("Missing search query")
		http.Error(w, "Missing search query", http.StatusBadRequest)

		return
	}

	members, err := s.repository.SearchMembers(query)
	if err != nil {
		fmt.Println("Error searching books:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)

		return
	}

	fmt.Printf("len(members): %v\n", len(members))
	fmt.Printf("members: %v\n", members)
	writeJSON(w, members)
}

func (s *Service) getMemberHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
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
		fmt.Println("Error reading request body:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	err = s.repository.AddMember(m)
	if err != nil {
		fmt.Println("Error adding member:", err)
		http.Error(w, fmt.Sprintf("Failed to add member: %v", err), http.StatusInternalServerError)

		return
	}

	fmt.Println("Successfully added member:", m)
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
	idStr := chi.URLParam(r, "id")

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
