package frontend

import (
	"errors"
	"net/http"
)

type ErrorPageData struct {
	Message string `json:"message"`
	Details string `json:"error"`
}

func (s *Service) errorPage(w http.ResponseWriter, msg string, err error) {
	data := ErrorPageData{
		Message: msg,
		Details: err.Error(),
	}
	s.executeTemplate(w, "error.gohtml", data)
}

func (s *Service) tempErrorPage(w http.ResponseWriter, r *http.Request) {
	s.errorPage(w, "Temporary Error", errors.New("This page is temporarily unavailable. Please try again later"))
}
