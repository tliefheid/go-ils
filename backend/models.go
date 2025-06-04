package backend

import "time"

// Book represents a library book
type Book struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Author          string `json:"author"`
	ISBN            string `json:"isbn"`
	Genre           string `json:"genre"`
	PublicationYear int    `json:"publication_year"`
	CopiesTotal     int    `json:"copies_total"`
	CopiesAvailable int    `json:"copies_available"`
}

// Member represents a library member
type Member struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Contact  string `json:"contact"`
	MemberID string `json:"member_id"`
}

// Borrowing represents a book borrowing record
type Borrowing struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	MemberID   int        `json:"member_id"`
	IssueDate  time.Time  `json:"issue_date"`
	DueDate    time.Time  `json:"due_date"`
	ReturnDate *time.Time `json:"return_date"`
	Fine       float64    `json:"fine"`
}
