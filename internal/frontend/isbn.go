package frontend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) isbnPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, "Failed to parse form data", err)
		return
	}

	isbn := r.FormValue("isbn")
	fmt.Printf("isbn: %v\n", isbn)
	// isbn := chi.URLParam(r, "isbn")
	if isbn == "" {
		s.errorPage(w, "Missing ISBN", errors.New("missing ISBN"))
		return
	}

	resp, err := http.Get(s.uri + "/isbn/" + isbn)
	if err != nil {
		s.errorPage(w, "Failed to fetch book info", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.executeTemplate(w, "isbn.gohtml", map[string]string{
			"Error": "Book not found",
		})

		return
	}

	var book model.Book

	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		s.errorPage(w, "Failed to decode book info", err)
		return
	}

	fmt.Printf("book: %+v\n", book)
	ctx := context.WithValue(r.Context(), "book", book)
	s.bookUpsertPage(w, r.WithContext(ctx))
}

func (s *Service) isbnPage(w http.ResponseWriter, r *http.Request) {
	s.executeTemplate(w, "isbn.gohtml", nil)
}
