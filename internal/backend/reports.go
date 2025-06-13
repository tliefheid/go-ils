package backend

import "net/http"

func (s *Service) getBorrowedBooksHandler(w http.ResponseWriter, r *http.Request) {
	data, err := s.repository.ListBorrowings()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, data)
}
