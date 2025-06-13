package frontend

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) returnPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("returnPost called")

	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing borrow ID", nil)
		return
	}

	resp, err := http.Post(s.uri+"/returns/"+id, "application/json", nil)
	if err != nil {
		s.errorPage(w, "Failed to return book", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		s.errorPage(w,
			"Failed to return book",
			errors.New("Invalid response status code: "+resp.Status),
		)

		return
	}

	s.borrowPage(w, r) // Redirect to borrow page after successful return
}
