package frontend

import "net/http"

func (s *Service) indexPage(w http.ResponseWriter, r *http.Request) {
	s.executeTemplate(w, "index.gohtml", nil)
}
