package postgres

import (
	"fmt"
	"time"

	"github.com/tliefheid/go-ils/internal/model"
)

func (s *Store) ListBorrowings() ([]model.BorrowingDetail, error) {
	rows, err := s.db.Query(`
	SELECT br.id, b.id, b.title, m.id, m.name, br.issue_date FROM borrowings br
	JOIN books b
	ON br.book_id = b.id
	JOIN members m
	ON br.member_id = m.id
	WHERE br.return_date IS NULL`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []model.BorrowingDetail

	for rows.Next() {
		var id, bookId, memberId int

		var title, name string

		var issueDate time.Time
		if err := rows.Scan(&id, &bookId, &title, &memberId, &name, &issueDate); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		bd := model.BorrowingDetail{
			ID:         id,
			BookID:     id,
			BookTitle:  title,
			MemberID:   memberId,
			MemberName: name,
			IssueDate:  issueDate,
		}
		result = append(result, bd)
	}

	return result, nil
}
func (s *Store) AddBorrowing(b model.Borrowing) error {
	var available int

	err := s.db.QueryRow("SELECT copies_available FROM books WHERE id=$1", b.BookID).Scan(&available)
	if err != nil {
		return err
	}

	if available < 1 {
		fmt.Println("No copies available for book ID:", b.BookID)
		return fmt.Errorf("no copies available for book ID: %d", b.BookID)
	}
	// Insert borrowing record
	issueDate := time.Now()

	_, err = s.db.Exec(`INSERT INTO borrowings (book_id, member_id, issue_date) VALUES ($1, $2, $3)`, b.BookID, b.MemberID, issueDate)
	if err != nil {
		fmt.Println("Error inserting borrowing record:", err)
		return err
	}

	// Update book inventory
	_, err = s.db.Exec("UPDATE books SET copies_available = copies_available - 1 WHERE id=$1", b.BookID)
	if err != nil {
		fmt.Println("Error updating book inventory:", err)
		return err
	}

	return nil
}
func (s *Store) GetBorrowing(id int) (*model.BorrowingDetail, error) {
	rows, err := s.db.Query(`
	SELECT br.id, b.id, b.title, m.id, m.name, br.issue_date FROM borrowings br
	JOIN books b
	ON br.book_id = b.id
	JOIN members m
	ON br.member_id = m.id
	WHERE br.id=$1`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*model.BorrowingDetail

	for rows.Next() {
		var id, bookId, memberId int

		var title, name string

		var issueDate time.Time
		if err := rows.Scan(&id, &bookId, &title, &memberId, &name, &issueDate); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		bd := &model.BorrowingDetail{
			ID:         id,
			BookID:     id,
			BookTitle:  title,
			MemberID:   memberId,
			MemberName: name,
			IssueDate:  issueDate,
		}
		result = append(result, bd)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("borrowing with id %d not found", id)
	}

	if len(result) > 1 {
		return nil, fmt.Errorf("multiple borrowings found with id %d", id)
	}

	return result[0], nil
}

func (s *Store) DeleteBorrowing(id int) error {
	_, err := s.db.Exec("DELETE FROM borrowings WHERE book_id=$1", id)
	if err != nil {
		fmt.Println("Error deleting borrowing:", err)
		return err
	}

	return nil
}

func (s *Store) ReturnBorrowing(id int) error {
	returnDate := time.Now()

	_, err := s.db.Exec(`UPDATE borrowings SET return_date=$1 WHERE id=$2`, returnDate, id)
	if err != nil {
		fmt.Println("Error updating borrowing return date:", err)
		return err
	}

	b, err := s.GetBorrowing(id)
	if err != nil {
		fmt.Println("Error getting borrowing details:", err)
		return err
	}

	fmt.Printf("b: %v\n", b)

	_, err = s.db.Exec("UPDATE books SET copies_available = copies_available + 1 WHERE id=$1", b.BookID)
	if err != nil {
		fmt.Println("Error updating book inventory after return:", err)
		return err
	}

	return nil
}
