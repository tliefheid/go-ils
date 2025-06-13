package postgres

import (
	"fmt"

	"github.com/tliefheid/go-ils/internal/model"
	"github.com/tliefheid/go-ils/internal/repository"
)

func (s *Store) ListBooks() ([]model.Book, error) {
	rows, err := s.db.Query("SELECT id, title, author, isbn, publication_year, copies_total, copies_available FROM books ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []model.Book

	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		books = append(books, b)
	}

	return books, nil
}

func (s *Store) SearchBookByISBN(isbn string) (*model.Book, error) {
	rows, err := s.db.Query(`SELECT id, title, author, isbn, publication_year, copies_total, copies_available FROM books WHERE isbn ILIKE '%' || $1 || '%'`, isbn)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var books []*model.Book

	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		books = append(books, &b)
	}

	fmt.Printf("search with isbn: len(books): %v\n", len(books))

	if len(books) == 0 {
		return nil, repository.ErrNotFound
	}

	if len(books) > 1 {
		return nil, fmt.Errorf("multiple books found with isbn %s", isbn)
	}

	return books[0], nil
}
func (s *Store) SearchBooks(search string) ([]model.Book, error) {
	rows, err := s.db.Query(`SELECT id, title, author, isbn, publication_year, copies_total, copies_available FROM books WHERE title ILIKE '%' || $1 || '%' OR author ILIKE '%' || $1 || '%' OR isbn ILIKE '%' || $1 || '%'`, search)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var books []model.Book

	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		books = append(books, b)
	}

	return books, nil
}

func (s *Store) AddBook(book model.Book) error {
	query := `INSERT INTO books (title, author, isbn, publication_year, copies_total, copies_available) VALUES ($1, $2, $3, $4, $5, $5) RETURNING id`

	err := s.db.QueryRow(query, book.Title, book.Author, book.ISBN, book.PublicationYear, book.CopiesTotal).Scan(&book.ID)
	if err != nil {
		return err
	}

	return nil
}
func (s *Store) GetBook(id int) (*model.Book, error) {
	rows, err := s.db.Query("SELECT id, title, author, isbn, publication_year, copies_total, copies_available FROM books WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []model.Book

	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		books = append(books, b)
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("book with id %d not found", id)
	}

	if len(books) > 1 {
		return nil, fmt.Errorf("multiple books found with id %d", id)
	}

	return &books[0], nil
}
func (s *Store) UpdateBook(book model.Book) error {
	query := `UPDATE books SET title=$1, author=$2, isbn=$3, publication_year=$4, copies_total=$5, copies_available=$6 WHERE id=$7`

	_, err := s.db.Exec(query, book.Title, book.Author, book.ISBN, book.PublicationYear, book.CopiesTotal, book.CopiesAvailable, book.ID)
	if err != nil {
		fmt.Println("Error updating book:", err)
		return err
	}

	return nil
}
func (s *Store) DeleteBook(id int) error {
	_, err := s.db.Exec("DELETE FROM books WHERE id=$1", id)
	if err != nil {
		fmt.Println("Error deleting book:", err)
		return err
	}

	return nil
}
