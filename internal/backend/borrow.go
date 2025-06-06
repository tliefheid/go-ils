package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) borrowBookHandler(w http.ResponseWriter, r *http.Request) {
	var req model.BorrowRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BookID == "" || req.MemberID == "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	bookID, err := strconv.Atoi(req.BookID)
	if err != nil || bookID <= 0 {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.Atoi(req.MemberID)
	if err != nil || memberID <= 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	borrow := model.Borrowing{
		BookID:    bookID,
		MemberID:  memberID,
		IssueDate: time.Now(),
	}

	err = s.repository.AddBorrowing(borrow)
	if err != nil {
		fmt.Println("Error adding borrowing:", err)
		http.Error(w, "Failed to borrow book", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) getBorrowingHandler(w http.ResponseWriter, r *http.Request) {
	b, err := s.repository.ListBorrowings()
	if err != nil {
		fmt.Println("Error fetching borrowings:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)

		return
	}

	writeJSON(w, b)
}

func (s *Service) getBorrowingDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid borrowing ID", http.StatusBadRequest)
		return
	}

	detail, err := s.repository.GetBorrowing(id)
	if err != nil {
		http.Error(w, "Borrowing not found", http.StatusNotFound)
		return
	}

	writeJSON(w, detail)
}
