package frontend

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/library-ils-backend/internal/model"
)

func (s *Service) reportsPage(w http.ResponseWriter, r *http.Request) {
	borrowedResp, err := http.Get(s.uri + "/reports/borrowed")
	if err != nil {
		http.Error(w, "Failed to fetch borrowed books", 500)
		return
	}

	defer borrowedResp.Body.Close()

	var borrowed []model.BorrowingDetail
	if err := json.NewDecoder(borrowedResp.Body).Decode(&borrowed); err != nil {
		http.Error(w, "Failed to decode borrowed books", 500)
		return
	}

	// overdueResp, err := http.Get(s.uri + "/reports/overdue")
	// if err != nil {
	// 	http.Error(w, "Failed to fetch overdue books", 500)
	// 	return
	// }

	// defer overdueResp.Body.Close()

	// var overdue []map[string]interface{}
	// if err := json.NewDecoder(overdueResp.Body).Decode(&overdue); err != nil {
	// 	http.Error(w, "Failed to decode overdue books", 500)
	// 	return
	// }

	data := map[string]interface{}{
		"Borrowed": borrowed,
		// "Overdue":  overdue,
	}

	s.executeTemplate(w, "reports.gohtml", data)
}
