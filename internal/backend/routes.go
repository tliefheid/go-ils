package backend

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Service) setupRoutes() {
	fmt.Println("setup routes")
	s.mux.Use(middleware.Logger)

	s.mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
	s.mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Not Found handler called")
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	s.mux.Get("/isbn/{isbn}", s.isbnInfoHandler)

	s.mux.Mount("/books", s.handleBooksRoutes())
	s.mux.Mount("/members", s.handleMemberRoutes())

	s.mux.Mount("/returns", s.handleReturnsRoutes())
	s.mux.Mount("/borrow", s.handleBorrowRoutes())
}

func (s *Service) handleReturnsRoutes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Post("/{id}", s.returnBookHandler)

	return mux
}

func (s *Service) handleBorrowRoutes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Post("/", s.borrowBookHandler)
	mux.Get("/", s.getBorrowingHandler)
	mux.Get("/{id}", s.getBorrowingDetailHandler)

	return mux
}

func (s *Service) handleBooksRoutes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", s.listBooksHandler)
	mux.Get("/search", s.searchBooks)
	mux.Post("/", s.addBookHandler)

	mux.Route("/{id}", func(mux chi.Router) {
		mux.Get("/", s.getBookHandler)
		mux.Put("/", s.editBookHandler)
		mux.Delete("/", s.deleteBookHandler)
	})

	return mux
}
func (s *Service) handleMemberRoutes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", s.listMembersHandler)
	mux.Post("/", s.addMemberHandler)

	mux.Route("/{id}", func(mux chi.Router) {
		mux.Get("/", s.getMemberHandler)
		mux.Put("/", s.editMemberHandler)
		mux.Delete("/", s.deleteMemberHandler)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
}
