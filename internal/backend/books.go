package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tliefheid/go-ils/internal/model"
	"github.com/tliefheid/go-ils/internal/repository"
)

func (s *Service) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := s.repository.ListBooks()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, books)
}

func (s *Service) searchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		fmt.Println("Missing search query")
		http.Error(w, "Missing search query", http.StatusBadRequest)

		return
	}

	books, err := s.repository.SearchBooks(query)
	if err != nil {
		fmt.Println("Error searching books:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)

		return
	}

	fmt.Printf("len(books): %v\n", len(books))
	fmt.Printf("books: %v\n", books)
	writeJSON(w, books)
}

func (s *Service) getBookHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing book ID", http.StatusBadRequest)
		return
	}
	// Convert id to int
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, err := s.repository.GetBook(idInt)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	writeJSON(w, book)
}

func (s *Service) isBookPresentHandler(w http.ResponseWriter, r *http.Request) {
	isbn := chi.URLParam(r, "isbn")
	if isbn == "" {
		http.Error(w, "Missing ISBN", http.StatusBadRequest)
		return
	}

	book, err := s.repository.SearchBookByISBN(isbn)
	if err != nil {
		if err == repository.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.Error(w, "Book not found, "+err.Error(), http.StatusNotFound)

		return
	}

	writeJSON(w, book)
}
func (s *Service) addBookHandler(w http.ResponseWriter, r *http.Request) {
	var b model.Book
	// parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &b); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.repository.AddBook(b)
	if err != nil {
		fmt.Printf("add book: repository err: %v\n", err)

		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)

		return
	}

	writeJSON(w, b)
}

func (s *Service) editBookHandler(w http.ResponseWriter, r *http.Request) {
	var b model.Book

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	if err := json.Unmarshal(body, &b); err != nil {
		fmt.Printf("unmarshal err: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	if b.ID < 0 {
		fmt.Println("Missing book ID: ", b.ID)
		http.Error(w, "Missing book ID", http.StatusBadRequest)

		return
	}

	fmt.Printf("update book: %+v\n", b)

	err = s.repository.UpdateBook(b)
	if err != nil {
		fmt.Printf("update book: repository err: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// First, delete all borrowings for this book (to avoid FK constraint errors)
	err = s.repository.DeleteBorrowing(id)
	if err != nil {
		http.Error(w, "Database error (borrowings)", http.StatusInternalServerError)
		return
	}

	err = s.repository.DeleteBook(id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
