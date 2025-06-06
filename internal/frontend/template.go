package frontend

import (
	"log"
	"net/http"
)

// ExecuteTemplate renders the named template with the provided data and writes to the http.ResponseWriter.
// tmpl: parsed *template.Template
func (s *Service) executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := s.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("template execution error: %v", err)
	}
}
