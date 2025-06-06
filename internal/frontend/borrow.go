package frontend

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) borrowPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.uri + "/borrow")
	if err != nil {
		http.Error(w, "Failed to fetch borrowings", 500)
		return
	}

	defer resp.Body.Close()

	var b []model.BorrowingDetail

	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		http.Error(w, "Failed to decode borrowings", 500)
		return
	}

	s.executeTemplate(w, "borrow.gohtml", b)
}

func (s *Service) borrowDetailsPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.uri + "/borrow/" + chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Failed to fetch borrowings", 500)
		return
	}

	defer resp.Body.Close()

	var b model.BorrowingDetail

	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		http.Error(w, "Failed to decode borrowings", 500)
		return
	}

	s.executeTemplate(w, "borrow_detail.gohtml", b)
}

func (s *Service) borrowPost(w http.ResponseWriter, r *http.Request) {
	bookID := r.FormValue("book_id")
	memberID := r.FormValue("member_id")

	if bookID == "" || memberID == "" {
		http.Error(w, "Missing fields", 400)
		return
	}

	payload := model.BorrowRequest{
		BookID:   bookID,
		MemberID: memberID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to create JSON payload", 500)
		return
	}

	resp, err := http.Post(s.uri+"/borrow", "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		http.Error(w, "Failed to borrow book", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
