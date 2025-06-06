package frontend

import (
	"fmt"
	"html/template"

	"github.com/go-chi/chi/v5"
)

type Service struct {
	mux  *chi.Mux
	uri  string
	tmpl *template.Template
}

type Config struct {
	BackendUri string
}

func New(cfg Config) (*Service, error) {
	s := new(Service)
	s.mux = chi.NewRouter()
	s.tmpl = template.Must(template.New("").ParseGlob("assets/*.gohtml"))
	s.uri = cfg.BackendUri
	s.setupRoutes()

	return s, nil
}

func (s *Service) Mux() *chi.Mux {
	fmt.Println("Returning mux")
	return s.mux
}
