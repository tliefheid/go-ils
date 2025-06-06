package model

import "time"

// Book represents a library book
type Book struct {
	ID              int    `json:"id"` // generated id
	Title           string `json:"title"`
	Author          string `json:"author"`
	ISBN            string `json:"isbn"`
	PublicationYear int    `json:"publication_year"`
	CopiesTotal     int    `json:"copies_total"`
	CopiesAvailable int    `json:"copies_available"`
}

// Member represents a library member
type Member struct {
	ID      int    `json:"id"` // generated id
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

// Borrowing represents a book borrowing record
type Borrowing struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	MemberID   int        `json:"member_id"`
	IssueDate  time.Time  `json:"issue_date"`
	ReturnDate *time.Time `json:"return_date,omitempty"` // nil if not returned
}

type BorrowingDetail struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	BookTitle  string     `json:"book_title"`
	MemberID   int        `json:"member_id"`
	MemberName string     `json:"member_name"`
	IssueDate  time.Time  `json:"issue_date"`
	ReturnDate *time.Time `json:"return_date,omitempty"` // nil if not returned
}
