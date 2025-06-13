package frontend

import (
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

	s.mux.Get("/", s.indexPage)
	s.mux.Get("/error", s.tempErrorPage)
	s.mux.Mount("/isbn", s.isbnRoutes())
	s.mux.Mount("/books", s.handleBooksRoutes())
	s.mux.Mount("/members", s.handleMembersRoutes())
	s.mux.Mount("/borrow", s.handleBorrowRoutes())
	s.mux.Mount("/return", s.handleReturnRoutes())
	s.mux.Get("/reports", s.reportsPage)
	// http.HandleFunc("/members", membersPage)
	// http.HandleFunc("/borrowed", borrowedBooksPage)
	// http.HandleFunc("/borrow", borrowBookHandler)
	// http.HandleFunc("/isbn-lookup", isbnLookupPage)
	// http.HandleFunc("/delete-book", deleteBookHandler)
	// http.HandleFunc("/update-book", updateBookHandler)
	// http.HandleFunc("/borrowed-detail", borrowedDetailPage)
	// http.HandleFunc("/return-borrowing", returnBorrowingHandler)
	// http.HandleFunc("/reports", reportsPage)
	// http.HandleFunc("/member-detail", memberDetailPage)
	// http.HandleFunc("/update-member", updateMemberHandler)
	// http.HandleFunc("/delete-member", deleteMemberHandler)
	// http.HandleFunc("/create-member", createMemberHandler)
	// s.mux.Get("/isbn/{isbn}", s.isbnInfoHandler)
	// s.mux.Mount("/members", s.handleMemberRoutes())
	// s.mux.Mount("/returns", s.handleReturnsRoutes())
	// s.mux.Mount("/borrow", s.handleBorrowRoutes())
}

func (s *Service) handleBooksRoutes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", s.booksPage)
	mux.Get("/{id}", s.bookDetailPage)
	mux.Post("/", s.bookPost)
	mux.Get("/upsert/{id}", s.bookUpsertPage)
	mux.Post("/delete/{id}", s.deleteBookPost)
	// mux.Post("/", s.addBookHandler)

	// mux.Route("/{id}", func(mux chi.Router) {
	// 	mux.Get("/", s.getBookHandler)
	// 	mux.Put("/", s.editBookHandler)
	// 	mux.Delete("/", s.deleteBookHandler)
	// })

	return mux
}
func (s *Service) handleMembersRoutes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", s.memberPage)
	mux.Get("/{id}", s.memberDetailPage)
	mux.Post("/", s.memberPost)
	mux.Post("/{id}/delete", s.memberDeletePost)

	// mux.Route("/{id}", func(mux chi.Router) {
	// 	mux.Get("/", s.getBookHandler)
	// 	mux.Put("/", s.editBookHandler)
	// 	mux.Delete("/", s.deleteBookHandler)
	// })

	return mux
}
func (s *Service) isbnRoutes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", s.isbnPage)
	mux.Post("/", s.isbnPost)

	return mux
}

func (s *Service) handleBorrowRoutes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Get("/", s.borrowPage)
	mux.Get("/{id}", s.borrowDetailsPage)
	mux.Post("/", s.borrowPost)

	return mux
}
func (s *Service) handleReturnRoutes() *chi.Mux {
	mux := chi.NewRouter()
	// mux.Get("/", s.borrowPage)
	mux.Post("/{id}", s.returnPost)
	// mux.Post("/", s.borrowPost)

	return mux
}
