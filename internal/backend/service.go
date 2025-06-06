package backend

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/repository"
)

type Service struct {
	mux        *chi.Mux
	repository repository.Store
}

type Config struct {
	Repository repository.Store
}

func New(cfg Config) (*Service, error) {
	s := new(Service)
	s.mux = chi.NewRouter()

	s.repository = cfg.Repository

	s.setupRoutes()

	return s, nil
}

func (s *Service) Mux() *chi.Mux {
	fmt.Println("Returning mux")
	return s.mux
}
