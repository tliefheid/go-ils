package frontend

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) booksPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var resp *http.Response

	var err error
	if q != "" {
		resp, err = http.Get(s.uri + "/books/search?q=" + q)
	} else {
		resp, err = http.Get(s.uri + "/books")
	}

	if err != nil {
		http.Error(w, "Failed to fetch books", 500)
		return
	}

	defer resp.Body.Close()

	var books []model.Book
	if err := json.NewDecoder(resp.Body).Decode(&books); err != nil {
		http.Error(w, "Failed to decode books", 500)
		return
	}

	s.executeTemplate(w, "books.gohtml", map[string]interface{}{
		"Books": books,
		"Query": q,
	})
}

func (s *Service) bookDetailPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing book id", 400)
		return
	}

	resp, err := http.Get(s.uri + "/books/" + id)
	if err != nil {
		http.Error(w, "Failed to fetch books", 500)
		return
	}

	defer resp.Body.Close()

	var book model.Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		http.Error(w, "Failed to decode books", 500)
		return
	}

	// Fetch members for borrow dropdown
	resp2, err := http.Get(s.uri + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}

	defer resp2.Body.Close()

	var members []model.Member
	if err := json.NewDecoder(resp2.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}

	s.executeTemplate(w, "book_detail.gohtml", struct {
		Book    *model.Book
		Members []model.Member
	}{&book, members})
}
