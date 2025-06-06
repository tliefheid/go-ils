package backend

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Service) returnBookHandler(w http.ResponseWriter, r *http.Request) {
	// var req struct {
	// 	BookID   int `json:"book_id"`
	// 	MemberID int `json:"member_id"`
	// }
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Invalid request", http.StatusBadRequest)
	// 	return
	// }
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing borrowing ID", http.StatusBadRequest)
		return
	}
	// Convert id to int
	borrowingID, err := strconv.Atoi(id)
	if err != nil || borrowingID <= 0 {
		http.Error(w, "Invalid borrowing ID", http.StatusBadRequest)
		return
	}

	// if err := json.Unmarshal(body, &req); err != nil {
	// 	http.Error(w, "Invalid JSON", http.StatusBadRequest)
	// 	return
	// }

	// if req.BookID == 0 || req.MemberID == 0 {
	// 	http.Error(w, "Missing fields", http.StatusBadRequest)
	// 	return
	// }
	// Find the latest unreturned borrowing
	// var borrowID int

	// var dueDate time.Time

	// err = s.db.QueryRow(`SELECT id, due_date FROM borrowings WHERE book_id=$1 AND member_id=$2 AND return_date IS NULL ORDER BY issue_date DESC LIMIT 1`, req.BookID, req.MemberID).Scan(&borrowID, &dueDate)
	// if err != nil {
	// 	http.Error(w, "No active borrowing found", http.StatusNotFound)
	// 	return
	// }
	err = s.repository.ReturnBorrowing(borrowingID)
	if err != nil {
		fmt.Println("Error returning borrowing:", err)
		http.Error(w, "Failed to return book", http.StatusInternalServerError)

		return
	}

	fmt.Println("Successfully returned borrowing with ID:", borrowingID)
	w.WriteHeader(http.StatusNoContent)
}
