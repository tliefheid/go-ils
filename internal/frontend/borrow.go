package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tliefheid/go-ils/internal/model"
)

func (s *Service) borrowPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.uri + "/borrow")
	if err != nil {
		s.errorPage(w, "failed to fetch borrowings", err)
		return
	}

	defer resp.Body.Close()

	var b []model.BorrowingDetail

	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		s.errorPage(w, "failed to decode borrowings", err)
		return
	}

	s.executeTemplate(w, "borrow.gohtml", b)
}

func (s *Service) borrowDetailsPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.uri + "/borrow/" + chi.URLParam(r, "id"))
	if err != nil {
		s.errorPage(w, "failed to fetch borrowing details", err)
		return
	}

	defer resp.Body.Close()

	var b model.BorrowingDetail

	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		s.errorPage(w, "failed to decode borrowing details", err)
		return
	}

	s.executeTemplate(w, "borrow_detail.gohtml", b)
}

func (s *Service) borrowPost(w http.ResponseWriter, r *http.Request) {
	bookID := r.FormValue("book_id")
	memberID := r.FormValue("member_id")

	if bookID == "" || memberID == "" {
		s.errorPage(w, "Missing book ID or member ID", nil)
		return
	}

	payload := model.BorrowRequest{
		BookID:   bookID,
		MemberID: memberID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.errorPage(w, "Failed to marshal borrow request", err)
		return
	}

	resp, err := http.Post(s.uri+"/borrow", "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		s.errorPage(w, "Failed to send borrow request", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)

		s.errorPage(w, "Failed to borrow book", errors.New(string(body)))

		return
	}

	s.borrowPage(w, r)
}
